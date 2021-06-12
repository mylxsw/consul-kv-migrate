package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/mylxsw/asteria/log"
	"github.com/pmezard/go-difflib/difflib"
)

var srcAddress, srcToken, targetAddress, targetToken, action string

func main() {
	flag.StringVar(&srcAddress, "src-addr", "source-ip:8500", "源 Consul 服务地址")
	flag.StringVar(&srcToken, "src-token", "", "源 Consul Token")
	flag.StringVar(&targetAddress, "target-addr", "target-ip:8500", "目标 Consul 服务地址")
	flag.StringVar(&targetToken, "target-token", "", "目标 Consul Token")
	flag.StringVar(&action, "action", "diff", "执行动作：diff-对比源和目标 Consul 配置差异，migrate-迁移源到目标并覆盖目标, 目标服务器的多余的 Key 不会被删除")

	flag.Parse()

	srcClient, err := consulapi.NewClient(&consulapi.Config{
		Scheme:  "http",
		Address: srcAddress,
		Token:   srcToken,
	})
	if err != nil {
		panic(err)
	}

	targetClient, err := consulapi.NewClient(&consulapi.Config{
		Scheme:  "http",
		Address: targetAddress,
		Token:   targetToken,
	})
	if err != nil {
		panic(err)
	}

	switch action {
	case "diff":
		diffSrcVsTarget(srcClient, targetClient)
	case "migrate":
		migrateSrcToTarget(srcClient, targetClient)
	default:
		panic("not support such action")
	}
}

type KV struct {
	Key   string
	Value string
}

type KVs []KV

func (kvs KVs) Len() int {
	return len(kvs)
}

func (kvs KVs) Less(i, j int) bool {
	return kvs[i].Key > kvs[j].Key
}

func (kvs KVs) Swap(i, j int) {
	kvs[i], kvs[j] = kvs[j], kvs[i]
}

func getSortedKvs(client *consulapi.Client) KVs {
	srcPairs, _, err := client.KV().List("config", nil)
	if err != nil {
		panic(err)
	}

	srcKvs := make(KVs, 0)
	for _, pair := range srcPairs {
		srcKvs = append(srcKvs, KV{Key: pair.Key, Value: string(pair.Value)})
	}

	sort.Sort(srcKvs)
	return srcKvs
}

func (kvs KVs) Keys() []string {
	keys := make([]string, 0)
	for _, k := range kvs {
		keys = append(keys, k.Key)
	}

	return keys
}

func (kvs KVs) Map() map[string]KV {
	keys := make(map[string]KV, 0)
	for _, k := range kvs {
		keys[k.Key] = k
	}

	return keys
}

func diffSrcVsTarget(srcClient, targetClient *consulapi.Client) {
	srcKvs := getSortedKvs(srcClient)
	targetKvs := getSortedKvs(targetClient)

	isSynced := true
	keyDiff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(strings.Join(srcKvs.Keys(), "\n")),
		B:        difflib.SplitLines(strings.Join(targetKvs.Keys(), "\n")),
		FromFile: "src-keys",
		ToFile:   "target-keys",
		Context:  0,
	}

	keyDiffText, err := difflib.GetUnifiedDiffString(keyDiff)
	if err != nil {
		panic(err)
	}

	if keyDiffText != "" {
		isSynced = false
		fmt.Printf("Keys 变更: \n")
		fmt.Println(keyDiffText)
		fmt.Println("------------------")
	}

	targetMap := targetKvs.Map()
	for _, src := range srcKvs {
		if target, ok := targetMap[src.Key]; ok {
			itemDiff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(src.Value),
				FromFile: "src",
				B:        difflib.SplitLines(target.Value),
				ToFile:   "target",
			}

			diffText, err := difflib.GetUnifiedDiffString(itemDiff)
			if err != nil {
				panic(err)
			}

			if diffText != "" {
				isSynced = false
				fmt.Printf("Key %s 的值变更:\n\n", src.Key)
				fmt.Printf("%s", diffText)
				fmt.Printf("\n-------------\n")
			}
		}
	}

	if isSynced {
		fmt.Println("同步结果检查：\n\n当前源服务和目标服务所有 Key 均一致")
	}
}

func migrateSrcToTarget(srcClient, targetClient *consulapi.Client) {
	pairs, meta, err := srcClient.KV().List("config", nil)
	if err != nil {
		panic(err)
	}

	log.With(meta).Infof("kv metas")
	for _, pair := range pairs {
		log.Debugf("sync %s", pair.Key)

		if _, err := targetClient.KV().Put(pair, nil); err != nil {
			panic(err)
		}
	}

	log.Info("同步完成")
	log.Infof("重新检查源服务和目标服务同步状态...")

	diffSrcVsTarget(srcClient, targetClient)
}
