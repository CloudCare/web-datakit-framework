## Web Datakit Framework (WDF) Changelog

### v1.0.2 - 2020/01/13

- Fix:
	- 修复了 wdf 获取到的 nsq 节点为 0 时，会触发 panic 的问题；
- Add:
	- 添加了 wdf 向 nsqlookup 多次查询节点的逻辑，当查询到的节点数为0时，会阻塞并等待200ms再次查询，如果连续3次查询到的节点数相同则结束查询。


### v1.0.1 - 2020/01/05

- Add
	-  项目第一次发布。
