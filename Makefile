all:
	$(shell cd cmd/chat; go build .)
	mv cmd/chat/chat owncom-chat
	$(shell cd cmd/console; go build .)
	mv cmd/console/console owncom-console
