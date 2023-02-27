# X-Ui

English Version of X-UI Panel with Updated Xray-Core, Multi-protocol & Multi-user, with more option and some of more advanced.

It's a fork from [HossinAsaadi's x-ui](https://github.com/hossinasaadi/x-ui) and [NidukaAkalanka's x-ui](https://github.com/NidukaAkalanka/x-ui-english).


## Features
-   The xray-core has been updated
-   Ability to install the latest version of the xary-core from inside the panel
-   More transmits for the Trojan protocol
-   Ability to define SNI for VMESS, VLESS, Trojan
-   Automatic Adding users email to configuration name by panel
-   Some design change for better looking
-   Improving the Persian language for the panel
-   Everything is in English (Serverside setup + Serverside UI + Web UI)
-   System status monitoring
-   Support multi-user multi-protocol, web page visualization operation
-   Multi UUIDs can be added as users for Vmess and Vless configurations with separate QR codes
-   Supported protocols: vmess, vless, trojan, shadowsocks, dokodemo-door, socks, http
-   Support to configure more transmission configurations
-   Traffic statistics, limit traffic, limit expiration time
-   Customizable xray configuration templates
-   Support https access panel (bring your own domain name + ssl certificate)
-   Support one-click install BBR kernel
-   Support one-click SSL certificate application and automatic renewal


## Supported operating systems

-   Debian 11 or higher (Recommended)
-   Ubuntu 20.04 or higher
-   Fedora 32 or higher
-   CentOS 9 Stream or higher
-   AlmaLinux 9.1 or higher
-   Rocky Linux 9 or higher

## Single Command Install & upgrade
Manual installation is recommended for **AlmaLinux** and **Rocky Linux**

    bash <(curl -Ls https://raw.githubusercontent.com/sudospaes/x-ui/master/install.sh)

    
## Manual install & upgrade
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
