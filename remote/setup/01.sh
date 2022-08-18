#!/bin/bash
set -eu

# ============================================================ #
# VARIABLES
# ============================================================ #

TIMEZONE=Asia/Jakarta

# the name of the new user to create
USERNAME=u_cinevie
DB_NAME=db_cinevie

# prompt to enter a password for the PostgreSQL cinevie user
# rather than hard-coding a password in this script
read -p "Enter password for cinevie DB user: " DB_PASSWORD

# force all output to be presented in en_US for duration of this script running
# this would avoid any "setting locale failed" errors while this script is running
export LC_ALL=en_US.UTF-8


# ============================================================= #
# LOGIC
# ============================================================= #

# enable the "universe" repository (package)
apt update
apt --yes install software-properties-common
apt update
add-apt-repository --yes universe

# update all software packages including configuration files if newer ones are available
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

# set the system timezone and install all locales
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# add the new user and give them sudo privileges
id -u "${USERNAME}" &> /dev/null || useradd --create-home --shell "/bin/bash" --groups sudo "${USERNAME}"

# force a password to be set for the new user the first time they log in
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"

# copy the SSH keys from root user to the new user
rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# configure firewall to allow SSH, HTTP, and HTTPS traffic
apt --yes install ufw
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# install fall2ban
apt --yes install fail2ban

# install the migrate CLI tool
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# install PostgreSQL
apt --yes install postgresql

# set up the cinevie DB and create a user account with the password entered earlier
# drop it particular element is exists
sudo -i -u postgres psql -c "DROP DATABASE IF EXISTS ${DB_NAME}"
sudo -i -u postgres psql -c "CREATE DATABASE ${DB_NAME}"

sudo -i -u postgres psql -d "${DB_NAME}" -c "CREATE EXTENSION IF NOT EXISTS citext"

sudo -i -u postgres psql -d "${DB_NAME}" -c "DROP ROLE IF EXISTS ${USERNAME}"
sudo -i -u postgres psql -d "${DB_NAME}" -c "CREATE ROLE ${USERNAME} WITH LOGIN PASSWORD '${DB_PASSWORD}'"

# add a DSN for connecting to the cinevie database to the system-wide environment
# variables in the /etc/environment file
echo "CINEVIE_DB_DSN='postgres://${USERNAME}:${DB_PASSWORD}@localhost/${DB_NAME}'" >> /etc/environment

# install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --batch --yes --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy -y
sudo apt autoremove -y

echo "Script complete! Rebooting..."
nohup reboot &>/dev/null & exit
