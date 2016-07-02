##排队服务器实现细则

###整体结构分为：
- 客户端Client
- 登录服务器LoginServer
- 排队服务器QueueServer

----


###执行流程：
1. Client向LoginServer发送登录请求。
2. LoginServer查检当前服务器负载，判断是否允许登录，负载过高，返回QueueServer的IP/PORT，并与LoginServer断开连接。
3. Client连接QueueServer，并保持PING/PONG请求。

----

###服务器实现原理：
1. #####登录服务器内部机制实现原理：#####
  - 如果LoginServer出现高负载情况，内部设定多长时间后允许下一个客户端登录。
  - LoginServer内部设定最大连接人数(即服务器承载人数上限)，负载人数(即人数达到负载人数后开始排队)。
  - LoginServer与QueueServer保持长连接，并间隔一定时间(30s)同步一次当前LoginServer当前负载情况。
  

2. #####排队服务器内部机制实现原理：#####
  - QueueServer与LoginServer保持长连接,每15s同步当前服务器负载。
  	- 如果LoginServer未达到承载人数上限，依据多长时间后允许下一个客户端登录的值，QueueServer计算客户端排队所需时间。
  	- 如果LoginServer已达到承载人数上限，QueueServer返回最长的排队时间。
  - Client连接QueueServer后，保持长连接，每15s同步当前服务器负载。
  - Client---------PING------->>>QueueServer
  - Client<<<------PONG----------QueueServer	(QS返回信息包含：客户端当前所处队列位置、排队所需多长时间。)
  - QueueServer内部维护一个正在排队的Client列表(FIFO)。
  

----
####可优化空间：
1. 涉及多次验证，Client登录高负载LoginServer会收到重定向到QueueServer的返回，并与LoginServer断开，
		高负载过后重新登录LoginServer。(优化：不做登录认证)

##安装和运行
1. 进入go项目
  - cd $GOPATH/src
2. 安装项目依赖beego/config模块
  - go get github.com/astaxie/beego
3.  下载serverFramework服务端框架
  - mkdir -p $GOPATH/src/bitbucket.org/serverFramework && cd $_
  - git clone https://zerak@bitbucket.org/serverFramework/serverFramework.git
4.  下载项目源码
  - git clone https://github.com/Zerak/VODone.git
    - 编译并运行QueueServer
      - cd $GOPATH/src/VODone/QueueServer
      - ./build
      - ./QueueServer
    - 编译并运行LoginServer
      - cd $GOPATH/src/VODone/LoginServer
      - ./build
      - ./LoginServer
    - 运行Client
      - cd $GOPATH/src/VODone/Client
      - go run main.go