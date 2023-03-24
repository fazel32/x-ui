# X-Ui

English Version of X-UI Panel with Updated Xray-Core, Multi-protocol & Multi-user, with more option and some of more advanced.

It's a fork from [HossinAsaadi's x-ui](https://github.com/hossinasaadi/x-ui) and [NidukaAkalanka's x-ui](https://github.com/NidukaAkalanka/x-ui-english).


## Features 🪄

-   System status monitoring
-   English and Persian language support
-   Use latest version of xray-core
-   Ability to install the latest version of the xray-core from inside the panel
-   Multi-user can be added as users for Vmess, Vless and Trojan configurations with separate QR codes
-   Traffic statistics, limit traffic, users tarffic limit and limit expiration time
-   Support more transmission for Trojan protocol
-   Ability to define SNI for Vmess, Vless, Trojan
-   Ability to define uTLS for Vmess, Vless, Trojan
-   Advanced TLS settings
-   Automatic Adding users email to configuration remark by panel
-   Some design change for better ui
-   Block Iran IP option
-   Security token for login to panel
-   Clone Inbounds
-   Support multi-user multi-protocol, web page visualization operation
-   Supported protocols: vmess, vless, trojan, shadowsocks, dokodemo-door, socks, http
-   Support to configure more transmission configurations
-   Customizable xray configuration templates
-   Support https access panel (bring your own domain name + ssl certificate)
-   Support one-click install BBR kernel
-   Support one-click SSL certificate application and automatic renewal


## Supported operating systems 💻
-   Debian 11 or higher (Recommended)
-   Ubuntu 20.04 or higher
-   Fedora 32 or higher
-   CentOS 9 Stream or higher
-   AlmaLinux 9.1 or higher
-   Rocky Linux 9 or higher

## Single Command Install & upgrade ✨

    bash <(curl -Ls https://raw.githubusercontent.com/sudospaes/x-ui/master/install.sh)

    
## Manual install & upgrade ⚙️
1.  First update your system and run the following commands. (Must have root user permissions)

   ```
sudo su
cd
```

2.  Then download the latest compressed package from  [https://github.com/sudospaes/x-ui/releases/latest](https://github.com/sudospaes/x-ui/releases/latest), generally choose  `amd64`  architecture

2.  Run the following commands respectively:

> If your server cpu architecture is not  `amd64`, replace  `*`  in the command with another architecture

  ```
rm x-ui/ /usr/local/x-ui/ /usr/bin/x-ui -rf
tar zxvf x-ui-linux-amd64.tar.gz
chmod +x x-ui/x-ui x-ui/bin/xray-linux-* x-ui/x-ui.sh
cp x-ui/x-ui.sh /usr/bin/x-ui
cp -f x-ui/x-ui.service /etc/systemd/system/
mv x-ui/ /usr/local/
systemctl daemon-reload
systemctl enable x-ui
systemctl restart x-ui
```
