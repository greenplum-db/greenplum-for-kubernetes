# Steps to run local docs server

## Pre-requisites
- git clone git@github.com:pivotal/gp-kubernetes-docs-book.git
- git clone git@github.com:pivotal-cf/docs-layout-repo.git
- git clone git@github.com:pivotal/greenplum-for-kubernetes.git
- Gemfile in `gp-kubernetes-docs-book` needs the following versions:
    ```
    gem 'bookbindery', '10.1.14'
    gem 'libv8', '3.16.14.15'
    gem 'nokogiri', '1.8.2'
    ```

## Create dependency image locally

This step takes few minutes to download required debian packages

```
cd greenplum-for-kubernetes/workstation-setup/local-docs-server
docker build -t local-docs-server:latest .
```

## Run Docker container

Now, run a container in detached mode to start a docs server as below:

```
docker run \
--rm \
-v ${HOME}/workspace/gp-kubernetes-docs-book:/gp-kubernetes-docs-book \
-v ${HOME}/workspace/greenplum-for-kubernetes:/greenplum-for-kubernetes \
-v ${HOME}/workspace/docs-layout-repo:/docs-layout-repo \
-p 9292:9292 \
-d local-docs-server:latest \
/greenplum-for-kubernetes/workstation-setup/local-docs-server/start-docs-server.sh
```

This will take few minutes to download `gem packages`.

Once the docs server is started, you can access it via `http://localhost:9292` on your browser.

## Reflect Changes

Watching for changes is currently not-supported (not sure how to do). 

To view the changes on the docs server, run the following:

```
docker exec -it <container-id> bash -c "cd /gp-kubernetes-docs-book; bundle exec bookbinder bind local"
```
Now, refresh your browser and see the changes
