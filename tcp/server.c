#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>

void error(const char *msg)
{
	perror(msg);
	exit(1);
}

int main(int argc, char *argv[])
{
	int n, sockfd, newsockfd, port;
	char buffer[256];
	socklen_t clilen;
	struct sockaddr_in serv_addr, cli_addr;

	if (argc < 2) {
		fprintf(stderr, "ERROR, no port provided\n");
		exit(1);
	}

	sockfd = socket(AF_INET, SOCK_STREAM, 0);
	if (sockfd < 0)
		error("ERROR opening socket");

    printf("Socket created\n");

	bzero((char *) &serv_addr, sizeof(serv_addr));
	port = atoi(argv[1]);

	serv_addr.sin_family = AF_INET;
	serv_addr.sin_addr.s_addr = INADDR_ANY;
	serv_addr.sin_port = htons(port);

	if (bind(sockfd, (struct sockaddr *) &serv_addr, sizeof(serv_addr)) < 0)
		error("ERROR on binding");

	listen(sockfd, 5);
	clilen = sizeof(cli_addr);

	newsockfd = accept(sockfd, (struct sockaddr *) &cli_addr, &clilen);
	if (newsockfd < 0)
		error("ERROR on accept");

    printf("Connection accepted\n");

	int flag = 1;
	if (-1 == setsockopt(newsockfd, SOL_SOCKET, SO_REUSEADDR, &flag, sizeof(flag))) {
		error("setsockopt fail");
	}

    printf("Connection on port %d\n", port);

	while(1) {
		bzero(buffer, 256);
		n = recv(newsockfd, buffer, 255, 0);
		if (n < 0)
			error("ERROR reading from socket");

		printf("Here is the message: %s\n", buffer);

		n = send(newsockfd, "pong", 5, 0);
		if (n < 0)
			error("ERROR writing to socket");
	}
	close(newsockfd);
	close(sockfd);
	return 0;
}