#! /bin/bash

# Verify git, process tools installed
apt-get update && apt-get -y install git procps ftp sqlite3 build-essential tzdata python-pip ffmpeg
pip install amcrest

curl -sL https://deb.nodesource.com/setup_14.x | bash - && apt-get install -y nodejs

curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg |  apt-key add - && \
    echo "deb https://dl.yarnpkg.com/debian/ stable main" |  tee /etc/apt/sources.list.d/yarn.list

npm install -g browserify watchify yarn

# Clean up
apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*s