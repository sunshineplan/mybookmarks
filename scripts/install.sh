#! /bin/bash

installSoftware() {
    apt -qq -y install nginx default-mysql-client
}

installMyBookmarks() {
    mkdir -p /var/www/mybookmarks
    curl -Lo- https://github.com/sunshineplan/mybookmarks/releases/download/v1.0/release.tar.gz | tar zxC /var/www/mybookmarks
    cd /var/www/mybookmarks
    chmod +x mybookmarks
}

configMyBookmarks() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter unix socket(default: /run/mybookmarks.sock): ' unix
    [ -z $unix ] && unix=/run/mybookmarks.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/mybookmarks.log): ' log
    [ -z $log ] && log=/var/log/app/mybookmarks.log
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/mybookmarks/config.ini.default > /var/www/mybookmarks/config.ini
    sed -i "s/\$header/$header/" /var/www/mybookmarks/config.ini
    sed -i "s/\$value/$value/" /var/www/mybookmarks/config.ini
    sed -i "s,\$unix,$unix," /var/www/mybookmarks/config.ini
    sed -i "s,\$log,$log," /var/www/mybookmarks/config.ini
    sed -i "s/\$host/$host/" /var/www/mybookmarks/config.ini
    sed -i "s/\$port/$port/" /var/www/mybookmarks/config.ini
    ./mybookmarks install
    service mybookmarks start
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
    cp -s /var/www/mybookmarks/scripts/mybookmarks.cron /etc/cron.monthly/mybookmarks
    chmod +x /var/www/mybookmarks/scripts/mybookmarks.cron
}

setupNGINX() {
    cp -s /var/www/mybookmarks/scripts/mybookmarks.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/mybookmarks/scripts/mybookmarks.conf
    sed -i "s,\$unix,$unix," /var/www/mybookmarks/scripts/mybookmarks.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installMyBookmarks
    configMyBookmarks
    writeLogrotateScrip
    createCronTask
    setupNGINX
}

main
