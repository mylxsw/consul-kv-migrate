# consul-kv-migrate

一个简单的小工具，用于在两个独立的 Consul 集群之间迁移 KV 数据和比较 KV 差异。

## 命令行参数

```bash
$ consul-kv-migrate -h
Usage of consul-kv-migrate:
  -action string
    	执行动作：diff-对比源和目标 Consul 配置差异，migrate-迁移源到目标并覆盖目标, 目标服务器的多余的 Key 不会被删除 (default "diff")
  -src-addr string
    	源 Consul 服务地址 (default "source-ip:8500")
  -src-token string
    	源 Consul Token
  -target-addr string
    	目标 Consul 服务地址 (default "target-ip:8500")
  -target-token string
    	目标 Consul Token
```

## 使用示例

将 192.168.1.10 的 Consul 集群上的所有 KV 同步到 192.168.2.10 的集群。

```bash
consul-kv-migrate \
    -target-addr 192.168.2.10:8500 \
    -target-token xxxxx-xxxxx-xxxxx-xxxx-xxxx \
    -src-addr 192.168.1.10:8500 \
    -src-token yyyy-yyyy-yyyyy-yyyyy-yyyy \
    -action migrate
```

对比 192.168.1.10 和 192.168.2.10 两个 Consul 集群的 KV 差异。

```bash
consul-kv-migrate \
    -target-addr 192.168.2.10:8500 \
    -target-token xxxxx-xxxxx-xxxxx-xxxx-xxxx \
    -src-addr 192.168.1.10:8500 \
    -src-token yyyy-yyyy-yyyyy-yyyyy-yyyy \
    -action diff
```


