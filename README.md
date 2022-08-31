# fake-jenkins

a fake implementation of the jenkins queue/build api to test remote triggering

## Run the local mock

`go run main.go`

## Run the Jenkins Adapter integration tests

`JENKINS_HOST_URL=http://127.0.0.1:8080/ JENKINS_USERNAME=jenkins JENKINS_API_TOKEN=token make int-test`

You should see an output like:

```
Ran 7 of 7 Specs in 1.750 seconds
SUCCESS! -- 7 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS

Ginkgo ran 1 suite in 3.786252458s
Test Suite Passed
```

## Using a self signed cert

Generate the cert with:

```
openssl req -nodes -new -x509 -keyout key.pem -out cert.pem -days 365 -subj '/CN=127.0.0.1'
```

Run the application with:

```
make darwin
./build/darwin/fake-jenkins -key ${HOME}/workspace/fake-jenkins/key.pem -cert ${HOME}/workspace/fake-jenkins/cert.pem
```

## Serving on localhost only

Add the `-local` option when running the program in order to server on localhost
(i.e.: 127.0.0.1) only.  This option will prevent the Mac OS firewall dialog box
from opening when running on Mac OS.

## Deploying to a cluster

```bash
ko resolve -f config/fake-jenkins-deployment.yml | kubectl apply -f -
```

## Deploying with a cert to a cluster

```bash
openssl req -nodes -new -x509 -keyout key.pem -out cert.pem -days 365 -subj '/CN=127.0.0.1'
ko resolve -f <(ytt -f config/ --data-value-file cert_pem=cert.pem --data-value-file key_pem=key.pem) | kubectl apply -f -
```
