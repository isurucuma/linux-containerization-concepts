# Dockerfile for Linux Containers Learning Environment
FROM ubuntu:22.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install essential tools
RUN apt-get update && apt-get install -y \
  build-essential \
  curl \
  wget \
  git \
  vim \
  nano \
  htop \
  tree \
  strace \
  gdb \
  valgrind \
  net-tools \
  iproute2 \
  iptables \
  bridge-utils \
  util-linux \
  procps \
  psmisc \
  lsof \
  tcpdump \
  netcat \
  telnet \
  socat \
  unshare \
  nsenter \
  cgroup-tools \
  systemd-container \
  docker.io \
  && rm -rf /var/lib/apt/lists/*

# Install Go
RUN curl -fsSL https://golang.org/dl/go1.21.0.linux-amd64.tar.gz | tar -xzC /usr/local
ENV PATH="/usr/local/go/bin:${PATH}"

# Create a working directory for learning materials
WORKDIR /learning

# Copy the learning materials from the host
COPY ../01-process-management /learning/01-process-management
COPY ../02-namespaces /learning/02-namespaces
COPY ../03-cgroups /learning/03-cgroups
COPY ../04-filesystem-isolation /learning/04-filesystem-isolation
COPY ../05-container-images /learning/05-container-images
COPY ../06-network-virtualization /learning/06-network-virtualization
COPY ../07-security-capabilities /learning/07-security-capabilities
COPY ../08-container-runtime /learning/08-container-runtime
COPY ../09-advanced-concepts /learning/09-advanced-concepts
COPY ../10-orchestration-basics /learning/10-orchestration-basics

# Make sure all demo files are executable
RUN find /learning -name "demo" -type f -exec chmod +x {} \;
RUN find /learning -name "*.go" -type f -exec chmod +r {} \;

# Create a user for the learning environment
RUN useradd -m -s /bin/bash learner && \
  echo 'learner:learner' | chpasswd && \
  usermod -aG sudo learner

# Set up environment variables
ENV USER=learner
ENV HOME=/home/learner
ENV SHELL=/bin/bash

# Create helpful aliases and functions
RUN echo 'alias ll="ls -la"' >> /home/learner/.bashrc && \
  echo 'alias demo="cd /learning && find . -name demo.go -type f"' >> /home/learner/.bashrc && \
  echo 'alias section="ls -la /learning/"' >> /home/learner/.bashrc && \
  echo 'export PS1="learner@container:\w# "' >> /home/learner/.bashrc

# Expose any necessary ports (for future networking sections)
EXPOSE 8080 9090

# Set the default user
USER learner

# Default command
CMD ["/bin/bash"]
