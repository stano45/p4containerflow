#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <time.h>

void error(const char *msg)
{
    perror(msg);
    exit(0);
}

int main(int argc, char *argv[])
{
    int sockfd, port, n, counter = 0;
    struct sockaddr_in serv_addr;
    struct hostent *server;
    char buffer[256];
    time_t start_time, current_time;
    double elapsed_time;

    if (argc < 3) {
        fprintf(stderr, "usage %s hostname port\n", argv[0]);
        exit(0);
    }

    port = atoi(argv[2]);

    sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0)
        error("ERROR opening socket");

    server = gethostbyname(argv[1]);
    if (!server) {
        fprintf(stderr,"ERROR, no such host\n");
        return -1;
    }

    bzero((char *) &serv_addr, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;

    bcopy((char *)server->h_addr, (char *)&serv_addr.sin_addr.s_addr, server->h_length);
    serv_addr.sin_port = htons(port);

    if (connect(sockfd, (struct sockaddr *) &serv_addr, sizeof(serv_addr)) < 0)
        error("ERROR connecting");

    start_time = time(NULL); // Record the start time

    while (1) {
        bzero(buffer, 256);
        sprintf(buffer, "ping-%d", counter++);

        n = send(sockfd, buffer, strlen(buffer), 0);
        if (n < 0)
            error("ERROR writing to socket");

        bzero(buffer, 256);
        n = recv(sockfd, buffer, 255, 0);
        if (n < 0)
            error("ERROR reading from socket");

        current_time = time(NULL); // Get the current time
        elapsed_time = difftime(current_time, start_time); // Calculate elapsed time

        printf("Time elapsed: %.0f seconds, Message: %s\n", elapsed_time, buffer);
        usleep(100000); // Sleep for 0.1 seconds
    }

    close(sockfd);
    return 0;
}
