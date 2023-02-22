# 简介

该脚本用于收集阿里云上的弹性公网IP(EIP)和负载均衡器(ALB、SLB)的信息并进行入库处理。如果需要收集其他信息可以进行相应的扩展。

# 运行

**注意：**只需要修改settings.yml文件，修改数据库信息以及阿里云账号信息即可运行。

## 1、下载压缩包

```shell
wget -c https://dl.google.com/go/go1.19.2.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```

## 2、设置 gopath 和 goroot

```shell
mkdir -pv ~/.go/src ~/.go/pkg ~/.go/bin
echo 'export GOROOT="/usr/local/go"' >> ~/.bash_profile
echo 'export GOPATH="$HOME/.go"' >> ~/.bash_profile
echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.bash_profile
```

注意

- **goroot：** go 的安装目录
- **gopath：**默认采用和 $GOROOT 一样的值，但从 Go 1.1 版本开始，你必须修改为其它路径。它可以包含多个包含 Go 语言源码文件、包文件和可执行文件的路径，而这些路径下又必须分别包含三个规定的目录：`src`、`pkg`和 `bin`，这三个目录分别用于存放源码文件、包文件和可执行文件。

## 3、设置代理

```shell
vim ~/.bash_profile         # 打开文件

export GO111MODULE=auto
export GOPROXY="https://goproxy.cn,https://goproxy.io,direct"
export GONOSUMDB="*"

source ~/.bash_profile  # 重启配置文件生效
```

## 4、如果你在linux，同时使用goland 和 vscore 两种编译器，建议把环境配置如下

```shell
sudo vim /etc/profile 

export GOROOT=/usr/local/go
export GO111MODULE=auto
export GOPROXY="https://goproxy.cn,https://goproxy.io,direct"
export GOPATH=$HOME/.go    #这是你的工程目录，需要手动创建
export PATH=$PATH:$GOROOT/bin

source /etc/profile   #执行该文件
```

## 5、运行脚本

先将脚本传到opt目录下

```shell
cd /opt/aliyunEipLb
go mod tidy
go build

#定时任务
cat /etc/cron.d/aliyunEipLb 
0 0 * * * root cd /opt/aliyunEipLb&&./aliyunEipLb
```



