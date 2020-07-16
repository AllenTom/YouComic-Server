<div align="center" >
    <p align="center" style="font-size:24px">
        Project Polaris | You Comic
    </p>
    <p>
        文件服务器中的漫画管理器部分，本仓库存放服务端API服务
    </p>
</div>

## YouComic Server

![](https://img.shields.io/badge/Project-Project%20Polaris-green)
![](https://img.shields.io/badge/Project-YouComic-green)
![](https://img.shields.io/badge/Version-1.0.0-yellow) 

YouComic Server 是整个YouComic的核心服务，提供了大量的内容管理接口，提供给图形化界面应用(Web\Supervisor\Desktop)使用

## 💻预览

黑色的控制台不需要截图😂

## 开发

服务端主要由Golang + Gin组成。

### 编译
release中的版本不一定是最新版本，可以根据需要自行编译需要的版本。需要一起添加的资源文件有

`./assets` 静态资源文件

`./config/setup.json` 启动参数

编译完成的程序目录大致为:
```
- assets
- config
    |
    - setup.json
- main.exe (主程序名称，会有不同)   
```

具体的编译方法详见[Go语言文档](https://golang.org/cmd/compile/)

可以使用Powershell 运行写好的`build.ps1`进行编译


### 🔗 链接
- [🏕️YouComic Blog](https://project-xpolaris.github.io/)
- [💻YouComic Studio](https://github.com/Project-XPolaris/YouComic-Studio)
- [🔨YouComic Supervisor](https://github.com/Project-XPolaris/YouComic-Supervisor)
- [🌐YouComic Web](https://github.com/Project-XPolaris/YouComic-Web)
- [📱YouComic Mobile Suit](https://github.com/Project-XPolaris/YouComic-Mobile-Suit)
- [⭐️Project Polaris](https://github.com/Project-XPolaris)
