#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#else
#include <unistd.h>
#endif

#include "curl/curl.h"

#include "./lib/cJSON/cJSON.h"

/* Constants */
#define OS_LINUX 1
#define OS_WIN 2
#define OS_MAC 3
#define ARCH_32 1
#define ARCH_64 2
#define OS_NAME_LINUX "linux"
#define OS_NAME_WIN "win"
#define OS_NAME_MAC "mac"
#define ARCH_NAME_32 "32"
#define ARCH_NAME_64 "64"

#define REQ_GET_TASK 1
#define REQ_GET_CRACK_INFO 2
#define REQ_SEND_RESULT 3
#define URI_GET_TASK "task"
#define URI_GET_CRACK_INFO "crack"
#define URI_SEND_RESULT "task"

static const char CONFIG_FILE[] = "d-bf.json", LOG_FILE[] = "d-bf.log";

/* Global variables */

char currentPath[PATH_MAX + 1], *urlApiVer;
int os, arch, sslVerify;

/* Functions forward declaration */

void setCurrentPath(void);
cJSON *getJsonObject(cJSON *object, const char *option, int exit);
cJSON *getJsonFile(void);
void setPlatform(void);
int getNumOfCpu();
void chkConfigs(void);
cJSON *getPlatform(void);
void setUrlApiVer(void);
const char *getReqUri(int req);
int sendRequest(int reqType, cJSON *data);
size_t processResponse(char *ptr, size_t size, size_t nmemb, int *reqType);
void reqGetTask(void);
void resGetTask(char *response);
void reqSendResult(void);
void resSendResult(char *response);
void reqGetCrackInfo(void);
void resGetCrackInfo(char *response);

/* Main function entry point */

int main(int argc, char **argv)
{
    /* Initialization */
    setCurrentPath();
    chkConfigs();
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

void setCurrentPath()
{
    // Linux
    readlink("/proc/self/exe", currentPath, PATH_MAX);
    dirname(currentPath);
    strcat(currentPath, "/");
}

cJSON *getJsonObject(cJSON *object, const char *option, int halt)
{
    cJSON *jsonBuf;

    jsonBuf = cJSON_GetObjectItem(object, option);

    if (!jsonBuf) {
        if (halt) {
            fprintf(stderr, "'%s' %s", option, "not found in config file!");
            exit(1);
        } else {
            return 0;
        }
    } else {
        return jsonBuf;
    }
}

cJSON *getJsonFile()
{
    char *strBuf;
    FILE *configFile;

    strBuf = (char*) malloc(PATH_MAX + 1);
    strcpy(strBuf, currentPath);
    strcat(strBuf, CONFIG_FILE);
    configFile = fopen(strBuf, "rb");
    free(strBuf);
    if (!configFile) {
        fprintf(stderr, "%s", "Config file not found!");
        exit(1);
    }

    long configLen;

    fseek(configFile, 0, SEEK_END);
    configLen = ftell(configFile);
    fseek(configFile, 0, SEEK_SET);
    strBuf = (char*) malloc(configLen + 1);
    fread(strBuf, 1, configLen, configFile);
    fclose(configFile);

    cJSON *jsonBuf;
    jsonBuf = cJSON_Parse(strBuf);
    free(strBuf);

    if (!jsonBuf) {
        fprintf(stderr, "%s", "Invalid JSON config file!");
        exit(1);
    }

    return jsonBuf;
}

void setPlatform()
{
    char platform[20], osName[6], archName[3];

    // Check OS
#if defined(_WIN32) || defined(_WIN64)
    os = OS_WIN;
    strcpy(osName, OS_NAME_WIN);
#elif defined(__linux__)
    os = OS_LINUX;
    strcpy(osName, OS_NAME_LINUX);
#elif defined(__APPLE__)
    os = OS_MAC;
    strcpy(osName, OS_NAME_MAC);
#endif

    // Check system type 32 or 64
    if (sizeof(void *) == 8) {
        arch = ARCH_64;
        strcat(archName, ARCH_NAME_64);
    } else {
        arch = ARCH_32;
        strcat(archName, ARCH_NAME_32);
    }

    // Add CPU
    strcpy(platform, osName);
    strcat(platform, "_");
    strcat(platform, archName);
    strcat(platform, "_cpu");

    cJSON *jsonArr, *jsonObj;
    jsonArr = cJSON_CreateArray();
    //TODO: Add all platforms in loop
    jsonObj = cJSON_CreateObject();
    cJSON_AddItemToObject(jsonObj, "id", cJSON_CreateString(platform));
    cJSON_AddItemToArray(jsonArr, jsonObj);

    cJSON *jsonBuf;
    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "platform", 0);
    if (jsonBuf) { // platform option exists in config file.
        jsonBuf = getJsonFile();
        cJSON_ReplaceItemInObject(jsonBuf, "platform", jsonArr);
    } else { // platform option does not exist in config file, so create it.
        jsonBuf = getJsonFile();
        cJSON_AddItemToObject(jsonBuf, "platform", jsonArr);
    }

    // Save new config file
    char *strBuf;
    FILE *configFile;
    strBuf = (char*) malloc(PATH_MAX + 1);
    strcpy(strBuf, currentPath);
    strcat(strBuf, CONFIG_FILE);
    configFile = fopen(strBuf, "wb");
    free(strBuf);
    if (!configFile) {
        fprintf(stderr, "%s", "Config file not found!");
        exit(1);
    }
    strBuf = cJSON_Print(jsonBuf);
    fputs(strBuf, configFile);
    fclose(configFile);

    free(strBuf);
    cJSON_Delete(jsonBuf);
    jsonObj = NULL;
    jsonArr = NULL;
    jsonBuf = NULL;
}

int getNumOfCpu(void)
{
    int pu = 0;
#if defined(_WIN32) || defined(_WIN64)
#ifndef _SC_NPROCESSORS_ONLN
    SYSTEM_INFO info;
    GetSystemInfo(&info);
#define sysconf(a) info.dwNumberOfProcessors
#define _SC_NPROCESSORS_ONLN
#endif
#endif
#ifdef _SC_NPROCESSORS_ONLN
    pu = sysconf(_SC_NPROCESSORS_ONLN);
#endif

    if (pu > 0)
        return pu;
    else
        return 0;
}

void chkConfigs(void)
{
    setPlatform();
}

cJSON *getPlatform()
{
    cJSON *jsonBuf;

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "platform", 1);
    return jsonBuf;
}

void setUrlApiVer()
{
    cJSON *jsonBuf;
    char strBuf[1025];
    size_t strSize;

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "server", 1);

    strcpy(strBuf, getJsonObject(jsonBuf, "url_api", 1)->valuestring);
    strSize = strlen(strBuf) + 3;
    urlApiVer = (char*) malloc(strSize);

    strcpy(urlApiVer, strBuf);
    strcat(urlApiVer, "/");
    strcpy(strBuf, getJsonObject(jsonBuf, "version", 1)->valuestring);
    strSize += strlen(strBuf);
    urlApiVer = (char*) realloc(urlApiVer, strSize);
    strcat(urlApiVer, strBuf);
    strcat(urlApiVer, "/");

    sslVerify = getJsonObject(jsonBuf, "ssl_verify", 1)->valueint;

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

int sendRequest(int reqType, cJSON *data)
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
        urlStr = (char*) malloc(
            strlen(urlApiVer) + strlen(getReqUri(reqType)) + 1);
        strcpy(urlStr, urlApiVer);
        strcat(urlStr, getReqUri(reqType));
        curl_easy_setopt(curl, CURLOPT_URL, urlStr);

        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, sslVerify);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, sslVerify);

        // Set post data
        char *strBuf;
        if (data) { // If the input json is valid
            strBuf = cJSON_PrintUnformatted(data);
            if (strBuf)
                curl_easy_setopt(curl, CURLOPT_POSTFIELDS, strBuf);
            else
                curl_easy_setopt(curl, CURLOPT_POSTFIELDS, "{}");
        } else {
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, "{}");
        }

        // Set callback for writing received data
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &reqType);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, processResponse);

        CURLcode resCode = curl_easy_perform(curl);
        curl_easy_cleanup(curl);

        free(urlStr);
        if (data)
            free(strBuf);

        return resCode;
    } else {
        return -1;
    }
}

size_t processResponse(char *ptr, size_t size, size_t nmemb, int *reqType)
{
    switch (*reqType) {
        case REQ_GET_TASK:
            resGetTask(ptr);
            break;
        case REQ_SEND_RESULT:
            resSendResult(ptr);
            break;
        case REQ_GET_CRACK_INFO:
            resGetCrackInfo(ptr);
            break;
    }

    return size * nmemb;
}

void reqGetTask(void)
{
    cJSON *jsonClientIfno = cJSON_CreateObject(), *jsonPlatform =
        cJSON_CreateObject();

    cJSON_AddItemToObject(jsonPlatform, "platform", getPlatform());
    cJSON_AddItemReferenceToObject(jsonClientIfno, "client_info", jsonPlatform);
    cJSON_PrintUnformatted(jsonClientIfno);

    sendRequest(REQ_GET_TASK, jsonClientIfno);
    cJSON_Delete(jsonClientIfno);
}

void resGetTask(char *response)
{
    cJSON *jsonBuf = cJSON_Parse(response);
    printf("response: %s", cJSON_PrintUnformatted(jsonBuf));
    cJSON_Delete(jsonBuf);
}

void reqSendResult(void)
{

}

void resSendResult(char *response)
{
    cJSON *jsonBuf = cJSON_Parse(response);

    cJSON_Delete(jsonBuf);
}

void reqGetCrackInfo(void)
{

}

void resGetCrackInfo(char *response)
{
    cJSON *jsonBuf = cJSON_Parse(response);

    cJSON_Delete(jsonBuf);
}
