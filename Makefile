all: miniRouterProto

miniRouterProto:
	$(MAKE) -C $@

clean:
	$(MAKE) -C proto $@

.PHONY: miniRouterProto