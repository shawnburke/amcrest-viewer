#-----------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See LICENSE in the project root for license information.
#-----------------------------------------------------------------------------------------

FROM golang:latest


# Copy default endpoint specific user settings overrides into container to specify Go path
COPY settings.vscode.json /root/.vscode-remote/data/Machine/settings.json

# Verify git, process tools installed
RUN apt-get update && apt-get -y install git procps ftp sqlite3



RUN curl -sL https://deb.nodesource.com/setup_14.x | bash - && apt-get install -y nodejs

RUN npm install -g browserify watchify

# Clean up
RUN apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*