#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/stat.h>

#include "curl/curl.h"

#include "./lib/cJSON/cJSON.h"
#include "./lib/base64/base64.h"

/* Determine OS */
#if defined(_WIN32) || defined(_WIN64)
#define OS_WIN
#elif defined(unix) || defined(__unix) || defined(__unix__) || defined(__linux__)
#define OS_LINUX
#elif defined(__APPLE__) || defined(__MACH__)
#define OS_MAC
#endif

/* Perform OS specific tasks */
#if defined(OS_WIN)
/* It is windows */
#include <windows.h>
#define WIN32_LEAN_AND_MEAN
#define PATH_SEPARATOR "\\"
#define OS_NAME "win"

#else
/* It is not windows */
#include <unistd.h>
#include <libgen.h>
#define PATH_SEPARATOR "/"

#endif

#if defined(OS_LINUX)
/* It is linux */
#define OS_NAME "linux"

#elif defined(OS_MAC)
/* It is mac */
#define OS_NAME "mac"

#endif

/* Constants */
#define ARCH_NAME_32 "32"
#define ARCH_NAME_64 "64"

#define REQ_GET_VENDOR 1
#define URI_GET_VENDOR "vendor/get"
#define REQ_UPDATE_VENDOR 2
#define URI_UPDATE_VENDOR "vendor/update"
#define REQ_GET_TASK 3
#define URI_GET_TASK "task/get"
#define REQ_GET_CRACK_INFO 4
#define URI_GET_CRACK_INFO "crack/info"
#define REQ_SEND_RESULT 5
#define URI_SEND_RESULT "task/result"

static const char CONFIG_FILE[] = "d-bf.json", LOG_FILE[] = "d-bf.log",
    RES_BODY_FILE[] = "res-body.tmp";

/* Global variables */

char archName[3], currentPath[PATH_MAX + 1], *urlApiVer;
int sslVerify;

/* Functions forward declaration */

int dirExists(const char *path);
int fileExists(const char *path);
void mkdirRecursive(char *path);
int fileGetContents(char **contents, const char *path, const char *errorMessage);
int fileCopy(const char *source, const char *target);
void setCurrentPath(void);
cJSON *getJsonObject(cJSON *object, const char *option, int exit);
cJSON *getJsonFile(void);
void setPlatform(void);
int getNumOfCpu(void);
void chkConfigs(void);
unsigned long long int benchmark(void);
cJSON *getPlatform(void);
void setUrlApiVer(void);
const char *getReqUri(int req);
size_t writeFunction(void *ptr, size_t size, size_t nmemb, void *stream);
int sendRequest(int reqType, cJSON *data);
void reqGetVendor(cJSON *vendorFile);
void resGetVendor(const char *resBodyPath, cJSON *vendorFile);
void reqUpdateVendor(void);
void resUpdateVendor(const char *resBodyPath);
void reqGetTask(void);
void resGetTask(const char *resBodyPath);
void reqSendResult(void);
void resSendResult(const char *resBodyPath);
void reqGetCrackInfo(void);
void resGetCrackInfo(const char *resBodyPath);

/* Main function entry point */

int main(int argc, char **argv)
{
    /* Initialization */
    setCurrentPath();
    setUrlApiVer();
    chkConfigs();

    // Global libcurl initialization
    if (curl_global_init(CURL_GLOBAL_ALL) != 0) {
        fprintf(stderr, "%s\n", "Error in initializing curl!");
        return 1; // Exit
    }

    // Destroy
    free(urlApiVer);
    curl_global_cleanup();

    return 0;
}

/* Functions definition */

int dirExists(const char *path)
{
    struct stat info;

    if (stat(path, &info) != 0)
        return 0;
    else
        if (info.st_mode & S_IFDIR)
            return 1;
        else
            return 0;
}

int fileExists(const char *path)
{
    struct stat st;
    return (stat(path, &st) == 0);
}

void mkdirRecursive(char *path)
{
    if (fileExists(path))
        return;

    char *subpath, *fullpath;

    fullpath = strdup(path);
    subpath = strdup(path);
    dirname(subpath);
    if (strlen(subpath) > 1)
        mkdirRecursive(subpath);
    mkdir(fullpath, S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH);
    free(fullpath);
}

int fileGetContents(char **contents, const char *path, const char *errorMessage)
{
    FILE *fileStream;

    fileStream = fopen(path, "rb");
    if (!fileStream) {
        if (errorMessage)
            fprintf(stderr, "%s\n", errorMessage);
        return -1;
    }

    long fileLength;

    fseek(fileStream, 0, SEEK_END);
    fileLength = ftell(fileStream);
    fseek(fileStream, 0, SEEK_SET);
    *contents = (char *) malloc(fileLength + 1);
    fread(*contents, 1, fileLength, fileStream);
    fclose(fileStream);

    return 0;
}

int fileCopy(const char *sourceFilePath, const char *targetFilePath)
{
    FILE *sourceStream, *targetStream;

    sourceStream = fopen(sourceFilePath, "rb");
    if (!sourceStream) {
        fprintf(stderr, "%s%s\n", "Can't read file: ", sourceFilePath);
    }

    char *targetDirName = strdup(targetFilePath);
    mkdirRecursive(dirname(targetDirName));
    free(targetDirName);
    targetStream = fopen(targetFilePath, "wb");
    if (!targetStream) {
        fclose(sourceStream);
        fprintf(stderr, "%s%s\n", "Can't write file: ", targetFilePath);
    }

    int ch;

    while ((ch = fgetc(sourceStream)) != EOF)
        fputc(ch, targetStream);

    fclose(sourceStream);
    fclose(targetStream);

    return 0;
}

void setCurrentPath(void)
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
        if (halt > -1) {
            fprintf(stderr, "'%s' %s\n", option, "not found in config file!");
            exit(halt);
        } else {
            return 0;
        }
    } else {
        return jsonBuf;
    }
}

cJSON *getJsonFile(void)
{
    char filePath[PATH_MAX + 1], *strBuf;

    strcpy(filePath, currentPath);
    strcat(filePath, CONFIG_FILE);

    if (fileGetContents(&strBuf, filePath, "Config file not found!") != 0)
        exit(1);

    cJSON *jsonBuf;
    jsonBuf = cJSON_Parse(strBuf);
    free(strBuf);

    if (!jsonBuf) {
        fprintf(stderr, "%s\n", "Invalid JSON config file!");
        exit(1);
    }

    return jsonBuf;
}

void setPlatform(void)
{
    char platformBase[9], platformId[20];

    // Check system type 32 or 64
    if (sizeof(void *) == 8) {
        strcat(archName, ARCH_NAME_64);
    } else {
        strcat(archName, ARCH_NAME_32);
    }

    strcpy(platformBase, OS_NAME);
    strcat(platformBase, "_");
    strcat(platformBase, archName);

    cJSON *jsonObj, *jsonArr = cJSON_CreateArray(), *vendorFile;
    char filePath[PATH_MAX + 1];

    // TODO: Add all platforms in loop
    strcpy(platformId, platformBase);
    strcat(platformId, "_cpu");
    jsonObj = cJSON_CreateObject();
    cJSON_AddItemToObject(jsonObj, "id", cJSON_CreateString(platformId));
    cJSON_AddItemReferenceToArray(jsonArr, jsonObj);

    // Check default vendor file
    strcpy(filePath, currentPath);
    strcat(filePath, "vendor");
    strcat(filePath, PATH_SEPARATOR);
    strcat(filePath, "cracker");
    strcat(filePath, PATH_SEPARATOR);
    strcat(filePath, "hashcat_cpu");
    strcat(filePath, PATH_SEPARATOR);
    strcat(filePath, platformId);

    if (!fileExists(filePath)) {
        vendorFile = cJSON_CreateObject();
        cJSON_AddItemToObject(vendorFile, "type",
            cJSON_CreateString("cracker"));
        cJSON_AddItemToObject(vendorFile, "name",
            cJSON_CreateString("hashcat_cpu"));
        cJSON_AddItemToObject(vendorFile, "platform_id",
            cJSON_CreateString(platformId));

        reqGetVendor(vendorFile);
        cJSON_Delete(vendorFile);
    }

    // Update platform in config file
    cJSON *jsonBuf;
    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "platform", -1);
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
    strBuf = (char *) malloc(PATH_MAX + 1);
    strcpy(strBuf, currentPath);
    strcat(strBuf, CONFIG_FILE);
    configFile = fopen(strBuf, "wb");
    free(strBuf);
    if (!configFile) {
        fprintf(stderr, "%s\n", "Config file not found!");
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
        return 1;
}

void chkConfigs(void)
{
    setPlatform();
}

unsigned long long int benchmark(void)
{
#ifdef CLOCK_PROCESS_CPUTIME_ID
    /* cpu time in the current process */
#define CLOCKTYPE  CLOCK_PROCESS_CPUTIME_ID
#else
    /* this one should be appropriate to avoid errors on multiprocessors systems */
#define CLOCKTYPE  CLOCK_MONOTONIC
#endif

    struct timespec ts, te;
    int i, cores = getNumOfCpu();
    long elaps_ns;
    double elapse;
    unsigned long long int hashes, benchmark = 0;
    char *b64Enc, b64Input[2] = "X";

    for (i = 0; i < cores; i++) {
        clock_gettime(CLOCKTYPE, &ts);
        hashes = 1;
        while (1) {
            b64Enc = (char *) malloc(Base64encode_len(strlen(b64Input)) + 1);
            Base64encode(b64Enc, b64Input, strlen(b64Input));
            free(b64Enc);

            clock_gettime(CLOCKTYPE, &te);
            elapse = difftime(te.tv_sec, ts.tv_sec);
            if (elapse > 0)
                break;
            hashes++;
        }

        elaps_ns = te.tv_nsec - ts.tv_nsec;
        elapse += ((double) elaps_ns) / 1.0e9;
        benchmark += (unsigned long long int) (hashes / elapse) + 0.5;
    }

    return benchmark;
}

cJSON *getPlatform(void)
{
    cJSON *jsonBuf;

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "platform", 1);
    return jsonBuf;
}

void setUrlApiVer(void)
{
    cJSON *jsonBuf;
    char strBuf[1025];

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "server", 1);

    strcpy(strBuf, getJsonObject(jsonBuf, "url_api", 1)->valuestring);
    strcat(strBuf, "/");
    strcat(strBuf, getJsonObject(jsonBuf, "version", 1)->valuestring);
    strcat(strBuf, "/");

    urlApiVer = (char *) malloc(strlen(strBuf) + 1);
    strcpy(urlApiVer, strBuf);

    sslVerify = getJsonObject(jsonBuf, "ssl_verify", 1)->valueint;

    cJSON_Delete(jsonBuf);
}

const char *getReqUri(int req)
{
    if (req == REQ_GET_VENDOR)
        return URI_GET_VENDOR;
    if (req == REQ_UPDATE_VENDOR)
        return URI_UPDATE_VENDOR;
    if (req == REQ_GET_TASK)
        return URI_GET_TASK;
    if (req == REQ_GET_CRACK_INFO)
        return URI_GET_CRACK_INFO;
    if (req == REQ_SEND_RESULT)
        return URI_SEND_RESULT;

    return "";
}

size_t writeFunction(void *ptr, size_t size, size_t nmemb, void *stream)
{
    size_t written = fwrite(ptr, size, nmemb, (FILE *) stream);
    return written;
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
        char urlStr[1025];

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
        char resBodyFilePath[PATH_MAX + 1];
        FILE *resBodyFile;

        strcpy(resBodyFilePath, currentPath);
        strcat(resBodyFilePath, RES_BODY_FILE);
        resBodyFile = fopen(resBodyFilePath, "wb");
        if (!resBodyFile) {
            curl_easy_cleanup(curl);

            fprintf(stderr, "%s\n", "Can't write response body file!");
            return -1;
        }

        curl_easy_setopt(curl, CURLOPT_WRITEDATA, resBodyFile);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeFunction);

        CURLcode resCode = curl_easy_perform(curl);
        curl_easy_cleanup(curl);
        fclose(resBodyFile);

        if (data && strBuf)
            free(strBuf);

        if (resCode == CURLE_OK) { // Request completed successfully, so process the response
            switch (reqType) {
                case REQ_GET_VENDOR:
                    resGetVendor(resBodyFilePath, data);
                    break;
                case REQ_UPDATE_VENDOR:
                    resUpdateVendor(resBodyFilePath);
                    break;
                case REQ_GET_TASK:
                    resGetTask(resBodyFilePath);
                    break;
                case REQ_SEND_RESULT:
                    resSendResult(resBodyFilePath);
                    break;
                case REQ_GET_CRACK_INFO:
                    resGetCrackInfo(resBodyFilePath);
                    break;
            }
        }

        // Delete temporary response file
        remove(resBodyFilePath);

        return resCode;
    } else {
        return -1;
    }
}

void reqGetVendor(cJSON *vendorFile)
{
    sendRequest(REQ_GET_VENDOR, vendorFile);
}

void resGetVendor(const char *resBodyPath, cJSON *vendorFile)
{
    char vendorFilePath[PATH_MAX + 1];

    strcpy(vendorFilePath, currentPath);
    strcat(vendorFilePath, "vendor");
    strcat(vendorFilePath, PATH_SEPARATOR);
    strcat(vendorFilePath, getJsonObject(vendorFile, "type", -1)->valuestring);
    strcat(vendorFilePath, PATH_SEPARATOR);
    strcat(vendorFilePath, getJsonObject(vendorFile, "name", -1)->valuestring);
    strcat(vendorFilePath, PATH_SEPARATOR);
    strcat(vendorFilePath,
        getJsonObject(vendorFile, "platform_id", -1)->valuestring);

    fileCopy(resBodyPath, vendorFilePath);

    chmod(vendorFilePath, S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH);
}

void reqUpdateVendor(void)
{
}

void resUpdateVendor(const char *resBodyPath)
{
}

void reqGetTask(void)
{
    cJSON *jsonReqData = cJSON_CreateObject(), *jsonPlatform =
        cJSON_CreateObject();

    cJSON_AddItemToObject(jsonPlatform, "platform", getPlatform());
    cJSON_AddItemToObject(jsonReqData, "client_info", jsonPlatform);

    sendRequest(REQ_GET_TASK, jsonReqData);
    cJSON_Delete(jsonReqData);
}

void resGetTask(const char *resBodyPath)
{
}

void reqSendResult(void)
{
}

void resSendResult(const char *resBodyPath)
{
}

void reqGetCrackInfo(void)
{
}

void resGetCrackInfo(const char *resBodyPath)
{
}
