# 示例说明  

## 前提条件 
* docker 容器  
    容器需要配置root用戶, 用于开启80端口权限{或者采用nginx代理80端口}   
    本地采用ubuntu环境  
  
属性 | 对应值    
--- | ---  
ip | 192.168.200.151 用于部署binlake-web     
user | root   
MySQL | 192.168.200.151:3358/root/secret
binlake-manager | http://192.168.200.152:9099/

* 安装      
    安装binlake-web-1.0.0  
    ```text
    * binlake-web编译  
    ```text
    # 下载 binlake-web 代码
    wget https://github.com/jd-tiger/binlake-web/releases/download/1.0.0/binlake-web-1.0.0.tar.gz
    
    # 解压到目标目录  
    tar -xf binlake-web-1.0.0.tar.gz -C /export/servers/binlake-web-1.0.0/
    ```
    

* 修改配置  
    * 初始化元数据库信息      
        ```text
        # 连接MySQL 初始化元数据信息  
        source ./conf/tower.sql   
        ```
    * 修改配置文件  
        * 修改MySQL 连接信息     
            ```text
            #### 元数据链接信息
            [meta]
            host = 192.168.200.151 {修改成本地ip}
            port = 3358
            user = root
            passwd = secret
            dbname = tower
            charset = utf8
            ```
        * 修改 dump 统一用户名密码  
            ```text
            ## binlog dump 统一的用户名和密码
            [dump]
            user = repl
            password = repl
            ```
            
        * 修改manager配置  
            ```text
            ## 管理端 api 地址 以及token
            [manager]
            url = http://192.168.200.152:9099/
            token = manager
            timeout = 120
            ```
            
        * 单点登录修改   
            采用统一的单点登录模块 
            ```text
            ## 单点登录
            [sso]
            url = http://test.ssa.jd.com/sso/ # 采用统一的单点登录模块 
            ```
  

* 启动binlake-web-1.0.0 
    注意: 这里需要开通80端口或者采用nginx代理方式    
    ```text
    ./restart.sh 
    ```
* 测试访问页面    
    注意: 如果是jd内部, 单点登录必须采用域名的方式访问并且后缀必须是 jd.com
    http://192.168.200.151  
    
...