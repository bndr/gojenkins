FROM jenkins/jenkins:2.138
RUN /usr/local/bin/install-plugins.sh cloudbees-folder ssh-slaves credentials ssh-credentials docker-commons
