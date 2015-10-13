#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "curl/curl.h"

#include "./lib/cJSON/cJSON.h"


/* Constants */

#define URI_GET_TASK "task"
#define URI_GET_CRACK_INFO "crack"
#define URI_SEND_RESULT "task"

#define REQ_GET_TASK 1
#define REQ_GET_CRACK_INFO 2
#define REQ_SEND_RESULT 3

static const char CONFIG_FILE[] = "d-bf.json", LOG_FILE[] = "d-bf.log";


/* Global variables */

char currentPath[PATH_MAX + 1], *urlApiVer, *bufferStr;
int sslVerify;
cJSON *bufferJson;


/* Functions forward declaration */

cJSON *getJsonObject(cJSON *object, const char *option);
cJSON *getJsonFile(void);
void setCurrentPath(void);
void checkConfigs(void);
void setUrlApiVer(void);
const char *getReqUri(int req);
int sendRequest(int reqType, const char *data);
size_t processResponse(char *ptr, size_t size, size_t nmemb, int *reqType);


/* Main function entry point */

int main(int argc, char **argv)
{
    /* Initialization */
    setCurrentPath();
    setUrlApiVer();

    // Global libcurl initialization
    if (curl_global_init(CURL_GLOBAL_ALL) != 0) {
        fprintf(stderr, "%s", "Error in initializing curl!");
        return 1; // Exit
    }

    // Destroy
    free(urlApiVer);
    curl_global_cleanup();

    return 0;
}


/* Functions definition */

cJSON *getJsonObject(cJSON *object, const char *option)
{
    cJSON *jsonBuf;

    jsonBuf = cJSON_GetObjectItem(object, option);

    if (!jsonBuf) {
        fprintf(stderr, "'%s' %s", option, "not found in config file!");
        exit(1);
    } else {
        return jsonBuf;
    }
}

cJSON *getJsonFile()
{
    FILE *configFile;

    bufferStr = (char*) malloc(PATH_MAX + 1);
    strncpy(bufferStr, currentPath, PATH_MAX);
    strcat(bufferStr, CONFIG_FILE);
    configFile = fopen(bufferStr, "rb");
    free(bufferStr);
    if (!configFile) {
        fprintf(stderr, "%s", "Config file not found!");
        exit(1);
    }

    long configLen;

    fseek(configFile, 0, SEEK_END);
    configLen = ftell(configFile);
    fseek(configFile, 0, SEEK_SET);
    bufferStr = (char*) malloc(configLen + 1);
    fread(bufferStr, 1, configLen, configFile);
    fclose(configFile);

    cJSON *jsonBuf;
    jsonBuf = cJSON_Parse(bufferStr);
    free(bufferStr);

    if (!jsonBuf) {
        fprintf(stderr, "%s", "Invalid JSON config file!");
        exit(1);
    }

    return jsonBuf;
}

void setCurrentPath()
{
    // Linux
    readlink("/proc/self/exe", currentPath, PATH_MAX);
    dirname(currentPath);
    strcat(currentPath, "/");
}

void checkConfigs(void)
{

}

void setUrlApiVer()
{
    cJSON *jsonBuf;
    char strBuf[1025];
    size_t strSize;

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "server");

    strcpy(strBuf, getJsonObject(jsonBuf, "url_api")->valuestring);
    strSize = strlen(strBuf) + 3;
    urlApiVer = (char*) malloc(strSize);
    strncpy(urlApiVer, strBuf, strlen(strBuf));
    strncat(urlApiVer, "/", 1);

    strcpy(strBuf, getJsonObject(jsonBuf, "version")->valuestring);
    strSize += strlen(strBuf);
    urlApiVer = (char*) realloc(urlApiVer, strSize);
    strncat(urlApiVer, strBuf, strlen(strBuf));
    strncat(urlApiVer, "/", 1);

    urlApiVer[strSize] = '\0';

    sslVerify = getJsonObject(jsonBuf, "ssl_verify")->valueint;

    cJSON_Delete(jsonBuf);
}

const char *getReqUri(int req)
{
    if (req == REQ_GET_TASK)
        return URI_GET_TASK;
    if (req == REQ_GET_CRACK_INFO)
        return URI_GET_CRACK_INFO;
    if (req == REQ_SEND_RESULT)
        return URI_SEND_RESULT;
}

int sendRequest(int reqType, const char *data)
{
    CURL *curl;

    curl = curl_easy_init();
    if (curl) {
        // Set request headers
        struct curl_slist *headers = NULL;
        headers = curl_slist_append(headers, "Accept: application/json");
        headers = curl_slist_append(headers, "Content-Type: application/json");
        headers = curl_slist_append(headers, "charsets: utf-8");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

        // Set request URL
        char *urlStr;
        urlStr = (char*) malloc(strlen(urlApiVer) + strlen(getReqUri(reqType)) + 1);
        strncpy(urlStr, urlApiVer, strlen(urlApiVer));
        strncat(urlStr, getReqUri(reqType), strlen(getReqUri(reqType)) + 1);
        curl_easy_setopt(curl, CURLOPT_URL, urlStr);
        free(urlStr);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, sslVerify);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, sslVerify);

        // Set post data
        cJSON *jsonBuf = cJSON_Parse(data);
        if (jsonBuf) // If the input json string is valid
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, data);
        else
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, "{}");
        cJSON_Delete(jsonBuf);

        // Set callback for writing received data
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, reqType);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, processResponse);

        CURLcode resCode = curl_easy_perform(curl);
        curl_easy_cleanup(curl);

        return resCode;
    } else {
        return -1;
    }
}

size_t processResponse(char *ptr, size_t size, size_t nmemb, int *reqType)
{
    cJSON *jsonBuf;

    jsonBuf = cJSON_Parse(ptr);
    printf("response: %s", cJSON_PrintUnformatted(jsonBuf));
    cJSON_Delete(jsonBuf);

    return size * nmemb;
}
