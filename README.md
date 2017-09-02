## lillian是什么?
lillian是用于管理外贸经济活动的CRM

### 技术支撑
* angularJs做前端
* go做后端
* 其中go使用net/http做支持，gorilla/mux做路由分发
* redis做缓存,mysql做数据存储
* session用到beego/session中的redis

### 部署,项目还没有完成，部署也不完善
* 先配置config目录下的config配置
* go build main.go
* 编译angularJs