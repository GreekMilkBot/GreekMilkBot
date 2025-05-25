# 关于消息路由

> 此为临时文档，完成后删除

## 目标

- [ ] 支持发送消息和接收消息
- [ ] 支持消息广播
- [ ] 消息可被序列化存储
- [ ] 消息具有优先级，部分类型的消息将被优先路由
- [ ] 支持使用 glob 接收一组消息
- [ ] 无法路由的消息将被安全的丢弃 （或者可以重放 ？）
- [ ] 支持 filter 拦截处理消息 （早于接收消息）

## 设计

简单的数据包结构如下:

```
type Packet struct{
    Src  string `json:"src"`
    Dest string `json:"dest"`
	
    Created time.Time `json:"created"`
    Content string `json:"content"`
	
}


```

所有的 `Dest` 默认使用 `<名称>/<路径>` 的形式, 其中:

- **名称**为分组名称
- **路径**为任意路径，支持子路径，不能以 `/` 开始和结束

