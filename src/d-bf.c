#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "curl/curl.h"

#include "./lib/cJSON/cJSON.h"

/* Constants */
#define REQ_GET_TASK 1
#define REQ_GET_CRACK_INFO 2
#define REQ_SEND_RESULT 3

static const char CONFIG_FILE[] = "d-bf.json";
static const char LOG_FILE[] = "d-bf.log";

/* Global variables */


/* Functions forward declaration */
int sendRequest(int reqType);
size_t processResponse(char *ptr, size_t size, size_t nmemb, int *reqType);

/* Main function entry point */
int main(int argc, char **argv)
{
	return 0;
    // Global libcurl initialisation
    if (curl_global_init(CURL_GLOBAL_ALL) != 0) {
        puts("cURL error!");
        return 1; // Exit
    }

    curl_global_cleanup();
    return 0;
}

/* Functions definition */

int sendRequest(int reqType)
{
    CURL *curl;
    CURLcode resCode;

    curl = curl_easy_init();
    if (curl) {
        // Set request headers
        struct curl_slist *headers = NULL;
        headers = curl_slist_append(headers, "Accept: application/json");
        headers = curl_slist_append(headers, "Content-Type: application/json");
        headers = curl_slist_append(headers, "charsets: utf-8");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

        // Set request URL
        curl_easy_setopt(curl, CURLOPT_URL, "");
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 0);

        // Set post data
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, "");

        // Set callback for writing received data
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, reqType);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, processResponse);

        resCode = curl_easy_perform(curl);
        curl_easy_cleanup(curl);

        return resCode;
    } else {
        return -1;
    }
}

size_t processResponse(char *ptr, size_t size, size_t nmemb, int *reqType)
{
    cJSON	*bufJson;

    bufJson = cJSON_Parse(ptr);
    printf("response: %s", cJSON_PrintUnformatted(bufJson));

    return size * nmemb;
}
