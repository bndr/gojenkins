FROM jenkins/jenkins:lts

COPY ./plugins.txt /tmp/plugins.txt
RUN jenkins-plugin-cli --plugin-file /tmp/plugins.txt
