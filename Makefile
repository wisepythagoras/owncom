all:
	$(shell cd cmd/chat; go build .)
	mv cmd/chat/chat owncom-chat
