#-----------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See LICENSE in the project root for license information.
#-----------------------------------------------------------------------------------------

FROM golang:latest


# Copy default endpoint specific user settings overrides into container to specify Go path
COPY settings.vscode.json /root/.vscode-remote/data/Machine/settings.json

COPY bootstrap.sh /root/bootstrap.sh
RUN /root/bootstrap.sh
