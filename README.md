msg-gate
=============

把smtp和企业微信发信封装为一个简单http接口，配置到open-falcon中用来发送报警信息

参考：

- [open-falcon/mail-provider](https://github.com/open-falcon/mail-provider)

- [ssl_smtp_example](https://gist.github.com/chrisgillis/10888032)

- [企业微信通知](https://github.com/Yanjunhui/chat)


## 安装方法

源码编译
下载之后为源码，安装golang环境，环境配置参考[golang环境配置](http://book.open-falcon.org/zh/quick_install/prepare.html)
编译方法
```bash
cd $GOPATH/src
mkdir github.com/wbzqe/ -p
cd github.com/wbzqe/
git clone https://github.com/wbzqe/msg-gate.git
cd msg-gate
go get ./...
./control build
```

使用下面命令打包为压缩包：
```bash
./control pack
```



## 使用方法
编译或者解压缩打包后的文件，修改cfg.json文件相关信息，使用下面命令启动
```bash
./control start
```
测试发信
```
curl http://$ip:4000/sender/mail -d "tos=a@a.com,b@b.com&subject=xx&content=yy"

curl http://$ip:4000/sender/qywx -d "tos=tanshuang&content=test"
```
