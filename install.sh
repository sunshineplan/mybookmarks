#! /bin/bash

installSoftware() {
    apt -qq -y install nginx
    apt -qq -y -t $(lsb_release -sc)-backports install golang-go
}

installMyIP() {
    curl -Lo- https://github.com/sunshineplan/mybookmarks-go/archive/v1.0.tar.gz | tar zxC /var/www
    mv /var/www/mybookmarks-go* /var/www/mybookmarks-go
    cd /var/www/mybookmarks-go
    go build
}

configMyIP() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter unix socket(default: /var/www/mybookmarks-go/mybookmarks-go.sock): ' unix
    [ -z $unix ] && unix=/var/www/mybookmarks-go/mybookmarks-go.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/mybookmarks-go.log): ' log
    [ -z $log ] && log=/var/log/app/mybookmarks-go.log
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/mybookmarks-go/config.ini.default > /var/www/mybookmarks-go/config.ini
    sed -i "s/\$header/$header/" /var/www/mybookmarks-go/config.ini
    sed -i "s/\$value/$value/" /var/www/mybookmarks-go/config.ini
    sed -i "s,\$unix,$unix," /var/www/mybookmarks-go/config.ini
    sed -i "s,\$log,$log," /var/www/mybookmarks-go/config.ini
    sed -i "s/\$host/$host/" /var/www/mybookmarks-go/config.ini
    sed -i "s/\$port/$port/" /var/www/mybookmarks-go/config.ini
}

setupsystemd() {
    cp -s /var/www/mybookmarks-go/mybookmarks-go.service /etc/systemd/system
    systemctl enable mybookmarks-go
    service mybookmarks-go start
}

writeLogrotateScrip() {
    if [ ! -f '/etc/logrotate.d/app' ]; then
	cat >/etc/logrotate.d/app <<-EOF
		/var/log/app/*.log {
		    copytruncate
		    rotate 12
		    compress
		    delaycompress
		    missingok
		    notifempty
		}
		EOF
    fi
}

createCronTask() {
    cp -s /var/www/mybookmarks-go/mybookmarks-go.cron /etc/cron.monthly
    chmod +x /var/www/mybookmarks-go/mybookmarks-go.cron
}

setupNGINX() {
    cp -s /var/www/mybookmarks-go/mybookmarks-go.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/mybookmarks-go/mybookmarks-go.conf
    sed -i "s,\$unix,$unix," /var/www/mybookmarks-go/mybookmarks-go.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installMyIP
    configMyIP
    setupsystemd
    writeLogrotateScrip
    createCronTask
    setupNGINX
}

main