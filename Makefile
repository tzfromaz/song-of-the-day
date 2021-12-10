local-build:
	go get .

local-run:
	go run .

docker-build:
	docker build -t tzfromaz/song-of-the-day .

docker-push:
	docker push tzfromaz/song-of-the-day

docker-run:
	docker run --name foobar -it --rm -p 30001:5000 tzfromaz/song-of-the-day

docker-stop:
	docker stop foobar

kube-deploy:
	kubectl apply -f foobar.yaml

kube-delete:
	kubectl delete -f foobar.yaml

eks-create-cluster:
	eksctl create cluster --name tweek --without-nodegroup --profile default

eks-create-nodegroup:
	eksctl create nodegroup \
      --cluster tweek \
      --region us-west-1 \
      --name tweek-nodegroup \
      --node-type t3.micro \
      --nodes 2 \
      --nodes-min 1 \
      --nodes-max 3 \
      --profile default

eks-delete:
	eksctl delete cluster --name tweek

twilio-config:
	$(eval HOSTNAME := $(shell kubectl get services bar --output jsonpath='{.status.loadBalancer.ingress[0].hostname}'))
	@curl -X POST https://api.twilio.com/2010-04-01/Accounts/${TWILIO_ACCOUNT_SID}/IncomingPhoneNumbers/${TWILIO_PHONE_NUMBER_SID}.json \
    --data-urlencode "SmsUrl=http://$(HOSTNAME)/sms" \
    -u ${TWILIO_ACCOUNT_SID}:${TWILIO_AUTH_TOKEN}