all: cmd

cmd: pb
	docker build --build-arg svrName=discovery -t discovery-svr .
	docker build --build-arg svrName=health -t health-svr .
	docker build --build-arg svrName=register -t register-svr .

pb:
	$(MAKE) -C $@

clean:
	$(MAKE) -C proto $@

.PHONY: pb