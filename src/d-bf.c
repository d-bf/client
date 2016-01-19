#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <ctype.h>
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
#define MAX_URL_LEN 2049

#define ARCH_NAME_32 "32"
#define ARCH_NAME_64 "64"

#define REQ_GET_VENDOR 1
#define URI_GET_VENDOR "vendor/get"
#define REQ_UPDATE_VENDOR 2
#define URI_UPDATE_VENDOR "vendor/update"
#define REQ_GET_ALGO_CRACKER 3
#define URI_GET_ALGO_CRACKER "get/algo-cracker"
#define REQ_GET_TASK 4
#define URI_GET_TASK "task/get"
#define REQ_GET_CRACK_INFO 5
#define URI_GET_CRACK_INFO "crack/info"
#define REQ_SEND_RESULT 6
#define URI_SEND_RESULT "task/result"

char CONFIG_PATH[8] = "config";
const char CONFIG_FILE[] = "d-bf.json", ALGO_CRACKER_DIR[] = "algo-cracker", CUSTOM_ALGO_CRACKER_EXT[] = ".force", LOG_FILE[] = "d-bf.log", RES_BODY_FILE[] = "res-body.tmp";

/* Global variables */

char archName[3], currentPath[PATH_MAX + 1], *urlApiVer;
int sslVerify;
cJSON *jsonBufTemp; // Used in doCrack()

/* Functions forward declaration */

void clearScreen(void);
void sleepSec(int seconds);
char *strReplace(char *str, char *find, char *rep);
char *getDirName(char *path);
int dirExists(const char *path);
int fileExists(const char *path);
void mkdirRecursive(char *path);
long int fileGetContents(char **contents, const char *path, const char *errorMessage);
int fileCopy(const char *source, const char *target);
void setCurrentPath(void);
void chkConfigs(void);
void syncPlatform(void);
cJSON *getJsonObject(cJSON *object, const char *option, const char *errorMessage);
cJSON *getJsonFile(void);
double getBench(char *benchStr);
long long int getBenchCpu(const char *vendorPath);
void chkPlatform(void);
cJSON *getPlatform(int active);
void setUrlApiVer(void);
void chkServerDependentConfigs(void);
const char *getReqUri(int req);
const char *getCracker(const char *mapFilePath, const char *algoId);
int doCrack(const char *crackInfoPath, cJSON **taskInfo);
size_t writeFunction(void *ptr, size_t size, size_t nmemb, void *stream);
int sendRequest(int reqType, cJSON *data);
void reqGetAlgoCracker(cJSON *reqData);
void resGetAlgoCracker(const char *resBodyPath, cJSON *reqData);
void reqGetVendor(cJSON *vendorFile);
void resGetVendor(const char *resBodyPath, cJSON *reqData);
void reqUpdateVendor(void);
void resUpdateVendor(const char *resBodyPath);
void reqGetTask(void);
void resGetTask(const char *resBodyPath);
void reqSendResult(cJSON *jsonResults);
void resSendResult(const char *resBodyPath, cJSON *reqData);
void reqGetCrackInfo(const char *crackId);
void resGetCrackInfo(const char *resBodyPath, cJSON *reqData);

/* Main function entry point */

int main(int argc, char **argv)
{
    clearScreen();
    printf("Initializing the program, please wait...\n\n");

    /* Initialization */
    setCurrentPath();
    chkConfigs();
    setUrlApiVer();
    chkServerDependentConfigs();

    // Global libcurl initialization
    if (curl_global_init(CURL_GLOBAL_ALL) != 0) {
        fprintf(stderr, "Error in initializing curl!\n");
        return 1; // Exit
    }

    clearScreen();

    int wait = 5; // Minimum time to wait between checks (in seconds)
    double elapsed;
    time_t startTime;
    while (1) { // TODO: Infinite loop
        startTime = time(NULL);

        printf("Checking for new task...\n");
        reqGetTask();
        printf("Done.\n\n");

        elapsed = difftime(time(NULL), startTime);
        if (elapsed < wait) {
            printf("Perform next check after a moment...\n\n");
            sleepSec((int) (wait - elapsed) + 1);
        }

        clearScreen();
    }

    // Destroy
    free(urlApiVer);
    curl_global_cleanup();

    return 0;
}

/* Functions definition */

void clearScreen(void)
{

}

void sleepSec(int seconds)
{
#if defined(OS_WIN)
    sleep(seconds * 1000);
#else
    sleep(seconds);
#endif
}

char *strReplace(char *str, char *find, char *rep)
{
    static char strReplaced[4096];
    static char strToReplace[4096];
    char *p;

    strncpy(strToReplace, str, 4095);
    strToReplace[4096] = '\0';
    strncpy(strReplaced, strToReplace, 4096);

    while ((p = strstr(strToReplace, find))) {
        strncpy(strReplaced, strToReplace, p - strToReplace); // Copy characters from 'str' start to 'orig' start
        strReplaced[p - strToReplace] = '\0';

        sprintf(strReplaced + (p - strToReplace), "%s%s", rep, p + strlen(find));

        strncpy(strToReplace, strReplaced, 4096);
    }

    return strReplaced;
}

char *getDirName(char *path)
{
    dirname(path);
    return path;
}

int dirExists(const char *path)
{
    struct stat info;

    if (stat(path, &info) != 0)
        return 0;
    else if (info.st_mode & S_IFDIR)
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

long int fileGetContents(char **contents, const char *path, const char *errorMessage)
{
    FILE *fileStream;

    fileStream = fopen(path, "rb");
    if (!fileStream) {
        if (errorMessage)
            fprintf(stderr, "%s\n", errorMessage);
        return -1;
    }

    long int fileLength;

    fseek(fileStream, 0, SEEK_END);
    fileLength = ftell(fileStream);
    fseek(fileStream, 0, SEEK_SET);
    *contents = (char *) malloc(fileLength + 1);
    fread(*contents, 1, fileLength, fileStream);
    fclose(fileStream);

    *(*contents + fileLength) = '\0';

    return fileLength;
}

int fileCopy(const char *sourceFilePath, const char *targetFilePath)
{
    FILE *sourceStream, *targetStream;

    sourceStream = fopen(sourceFilePath, "rb");
    if (!sourceStream) {
        fprintf(stderr, "Can't read file: %s\n", sourceFilePath);
    }

    char *targetDirName = strdup(targetFilePath);
    mkdirRecursive(dirname(targetDirName));
    free(targetDirName);
    targetStream = fopen(targetFilePath, "wb");
    if (!targetStream) {
        fclose(sourceStream);
        fprintf(stderr, "Can't write file: %s\n", targetFilePath);
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

void chkConfigs(void)
{
    // Create config path directory
    char configFilePath[PATH_MAX + 1];
    strcpy(configFilePath, currentPath);
    strcat(configFilePath, CONFIG_PATH);
    mkdirRecursive(configFilePath);

    strcat(CONFIG_PATH, PATH_SEPARATOR);

    strcat(configFilePath, PATH_SEPARATOR);
    strcat(configFilePath, CONFIG_FILE);

    if (fileExists(configFilePath)) {
        syncPlatform();
    } else {
        FILE *configFile;
        configFile = fopen(configFilePath, "wb");
        if (!configFile) {
            fprintf(stderr, "Can not create config file!\n");
            exit(1);
        }

        cJSON *jsonConfig = cJSON_CreateObject(), *jsonBuf = cJSON_CreateObject();
        cJSON_AddItemToObject(jsonBuf, "url_api", cJSON_CreateString(""));
        cJSON_AddItemToObject(jsonBuf, "version", cJSON_CreateString("v1"));
        cJSON_AddItemToObject(jsonBuf, "ssl_verify", cJSON_CreateNumber(0));
        cJSON_AddItemReferenceToObject(jsonConfig, "server", jsonBuf);

        char *strBuf = cJSON_Print(jsonConfig);
        fputs(strBuf, configFile);
        fclose(configFile);

        free(strBuf);
        cJSON_Delete(jsonConfig);

        syncPlatform();

        fprintf(stdout, "Please enter server's URL in url_api in config file: %s\n", configFilePath);
        exit(1);
    }
}

void syncPlatform(void)
{
    // Check if platform option exists in config file
    jsonBufTemp = getJsonFile();
    jsonBufTemp = getJsonObject(jsonBufTemp, "platform", NULL);
    if (jsonBufTemp) // platform option exists in config file
        return;

    // Check system type 32 or 64
    if (sizeof(void *) == 8) {
        strcat(archName, ARCH_NAME_64);
    } else {
        strcat(archName, ARCH_NAME_32);
    }

    char platformBase[9], platformId[20];

    strcpy(platformBase, OS_NAME);
    strcat(platformBase, "_");
    strcat(platformBase, archName);

    cJSON *jsonBuf, *jsonArr = cJSON_CreateArray();

    // Add CPU
    strcpy(platformId, "cpu_");
    strcat(platformId, platformBase);
    jsonBuf = cJSON_CreateObject();
    cJSON_AddItemToObject(jsonBuf, "id", cJSON_CreateString(platformId));
    cJSON_AddItemToObject(jsonBuf, "active", cJSON_CreateNumber(1));
    cJSON_AddItemToObject(jsonBuf, "benchmark", cJSON_CreateNumber(0));
    cJSON_AddItemReferenceToArray(jsonArr, jsonBuf);

    // Add GPU AMD
    strcpy(platformId, "gpu_");
    strcat(platformId, platformBase);
    strcat(platformId, "_amd");
    jsonBuf = cJSON_CreateObject();
    cJSON_AddItemToObject(jsonBuf, "id", cJSON_CreateString(platformId));
    cJSON_AddItemToObject(jsonBuf, "active", cJSON_CreateNumber(0));
    cJSON_AddItemToObject(jsonBuf, "benchmark", cJSON_CreateNumber(0));
    cJSON_AddItemReferenceToArray(jsonArr, jsonBuf);

    // Add GPU NVIDIA
    strcpy(platformId, "gpu_");
    strcat(platformId, platformBase);
    strcat(platformId, "_nv");
    jsonBuf = cJSON_CreateObject();
    cJSON_AddItemToObject(jsonBuf, "id", cJSON_CreateString(platformId));
    cJSON_AddItemToObject(jsonBuf, "active", cJSON_CreateNumber(0));
    cJSON_AddItemToObject(jsonBuf, "benchmark", cJSON_CreateNumber(0));
    cJSON_AddItemReferenceToArray(jsonArr, jsonBuf);

    // Add platform to config file
    jsonBufTemp = getJsonFile();
    cJSON_AddItemToObject(jsonBufTemp, "platform", jsonArr);

    // Save new config file
    char *strBuf;
    FILE *configFile;
    strBuf = (char *) malloc(PATH_MAX + 1);
    strcpy(strBuf, currentPath);
    strcat(strBuf, CONFIG_PATH);
    strcat(strBuf, CONFIG_FILE);
    configFile = fopen(strBuf, "wb");
    free(strBuf);
    if (!configFile) {
        fprintf(stderr, "Config file not found!\n");
        exit(1);
    }
    strBuf = cJSON_Print(jsonBufTemp);
    fputs(strBuf, configFile);
    fclose(configFile);

    free(strBuf);
    cJSON_Delete(jsonBuf);
}

cJSON *getJsonObject(cJSON *object, const char *option, const char *errorMessage)
{
    cJSON *jsonBuf;

    jsonBuf = cJSON_GetObjectItem(object, option);

    if (!jsonBuf) {
        if (errorMessage) {
            fprintf(stderr, "%s\n", errorMessage);
        }

        return NULL;
    } else {
        return jsonBuf;
    }
}

cJSON *getJsonFile(void)
{
    char filePath[PATH_MAX + 1], *strBuf;

    strcpy(filePath, currentPath);
    strcat(filePath, CONFIG_PATH);
    strcat(filePath, CONFIG_FILE);

    if (fileGetContents(&strBuf, filePath, "Config file not found!") < 0)
        exit(1);

    cJSON *jsonBuf;
    jsonBuf = cJSON_Parse(strBuf);
    free(strBuf);

    if (!jsonBuf) {
        fprintf(stderr, "Invalid JSON config file!\n");
        exit(1);
    }

    return jsonBuf;
}

double getBench(char *benchStr)
{
    double bench = 0;

    while (*benchStr) {
        if (isdigit(*benchStr)) {
            bench = strtod(benchStr, &benchStr);
            break;
        } else {
            benchStr++;
        }
    }

    if ((*benchStr == ' ')) {
        bench /= 1048576;
    } else if ((*benchStr == 'k') || (*benchStr == 'K')) { // Kilo
        bench /= 1024;
    } else if ((*benchStr == 'g') || (*benchStr == 'G')) { // Giga
        bench *= 1024;
    } else if ((*benchStr == 't') || (*benchStr == 'T')) { // Tera
        bench *= 1048576;
    } else if ((*benchStr == 'p') || (*benchStr == 'P')) { // Peta
        bench *= 1073741824;
    } else if ((*benchStr == 'e') || (*benchStr == 'E')) { // Exa
        bench *= 1099511627776;
    } else if ((*benchStr == 'z') || (*benchStr == 'Z')) { // Zetta
        bench *= 1125899906842624;
    } else if ((*benchStr == 'y') || (*benchStr == 'Y')) { // Yotta
        bench *= 1152921504606846976;
    }

    return bench;
}

long long int getBenchCpu(const char *vendorPath)
{
    char benchStr[256], cmdBench[PATH_MAX + 1];
    FILE *benchStream;
    double bench = 0;

    strcpy(cmdBench, vendorPath);
    strcat(cmdBench, " -b -m0");
    if (!(benchStream = popen(cmdBench, "r"))) {
        fprintf(stderr, "Can't get benchmark: %s\n", cmdBench);
        exit(1);
    }
    while (fgets(benchStr, sizeof(benchStr) - 1, benchStream) != NULL) {
        if (strncmp(benchStr, "Speed/sec: ", 11) == 0) {
            bench += getBench(benchStr);
        }
    }
    pclose(benchStream);

    return llround(bench / 1);
}

void chkPlatform(void)
{
    // Get platform from config file to update benchmarks
    cJSON *jsonConfig = getJsonFile();
    cJSON *jsonBuf = jsonConfig;
    jsonBuf = getJsonObject(jsonBuf, "platform", "'platform' not found in config file!");
    if (!jsonBuf)
        exit(1);

    char vendorPath[PATH_MAX + 1];
    cJSON *vendorData = NULL;

    jsonBuf = jsonBuf->child;
    while (jsonBuf) { // For each active platform
        jsonBufTemp = getJsonObject(jsonBuf, "active", "'active' not found in 'platform' in config file!");
        if (!jsonBufTemp || (jsonBufTemp->valueint == 0)) { // Is not active
            jsonBuf = jsonBuf->next;

            continue;
        }

        jsonBufTemp = getJsonObject(jsonBuf, "id", "'id' not found in 'platform' in config file!");
        if (jsonBufTemp) {
            // Check default vendor info
            strcpy(vendorPath, currentPath);
            strcat(vendorPath, "vendor");
            strcat(vendorPath, PATH_SEPARATOR);
            strcat(vendorPath, "cracker");
            strcat(vendorPath, PATH_SEPARATOR);
            strcat(vendorPath, "hashcat");
            strcat(vendorPath, PATH_SEPARATOR);
            strcat(vendorPath, jsonBufTemp->valuestring); // Platform id
            strcat(vendorPath, PATH_SEPARATOR);
            strcat(vendorPath, "config.json");
            if (!fileExists(vendorPath)) {
                vendorData = cJSON_CreateObject();
                cJSON_AddItemToObject(vendorData, "object_type", cJSON_CreateString("info"));
                cJSON_AddItemToObject(vendorData, "vendor_type", cJSON_CreateString("cracker"));
                cJSON_AddItemToObject(vendorData, "name", cJSON_CreateString("hashcat"));
                cJSON_AddItemToObject(vendorData, "platform_id", cJSON_CreateString(jsonBufTemp->valuestring)); // Platform id

                reqGetVendor(vendorData);
            }
            if (!fileExists(vendorPath)) {
                if (vendorData)
                    cJSON_Delete(vendorData);
                fprintf(stderr, "Vendor info not gotten: %s\n", vendorPath);
                exit(1);
            }

            // Check default vendor file
            strcpy(vendorPath, getDirName(vendorPath));
            strcat(vendorPath, PATH_SEPARATOR);
            strcat(vendorPath, jsonBufTemp->valuestring); // Platform id
            if (!fileExists(vendorPath)) {
                if (vendorData) {
                    strcpy(getJsonObject(vendorData, "object_type", NULL)->valuestring, "file");
                } else {
                    vendorData = cJSON_CreateObject();
                    cJSON_AddItemToObject(vendorData, "object_type", cJSON_CreateString("file"));
                    cJSON_AddItemToObject(vendorData, "vendor_type", cJSON_CreateString("cracker"));
                    cJSON_AddItemToObject(vendorData, "name", cJSON_CreateString("hashcat"));
                    cJSON_AddItemToObject(vendorData, "platform_id", cJSON_CreateString(jsonBufTemp->valuestring)); // Platform id
                }

                reqGetVendor(vendorData);
            }
            if (!fileExists(vendorPath)) {
                if (vendorData)
                    cJSON_Delete(vendorData);
                fprintf(stderr, "Vendor file not downloaded: %s\n", vendorPath);
                exit(1);
            }

            // Update benchmark
            jsonBufTemp = getJsonObject(jsonBuf, "benchmark", NULL);
            if (jsonBufTemp) { // benchmark option exists in platform in config file
                cJSON_ReplaceItemInObject(jsonBuf, "benchmark", cJSON_CreateNumber(getBenchCpu(vendorPath)));
            } else { // benchmark option does not exist in platform in config file
                cJSON_AddItemToObject(jsonBuf, "benchmark", cJSON_CreateNumber(getBenchCpu(vendorPath)));
            }
        }

        jsonBuf = jsonBuf->next;
    }

    // Save new config file
    char *strBuf;
    FILE *configFile;
    strBuf = (char *) malloc(PATH_MAX + 1);
    strcpy(strBuf, currentPath);
    strcat(strBuf, CONFIG_PATH);
    strcat(strBuf, CONFIG_FILE);
    configFile = fopen(strBuf, "wb");
    free(strBuf);
    if (!configFile) {
        fprintf(stderr, "Config file not found!\n");
        exit(1);
    }
    strBuf = cJSON_Print(jsonConfig);
    fputs(strBuf, configFile);
    fclose(configFile);

    free(strBuf);
    cJSON_Delete(jsonBuf);
}

cJSON *getPlatform(int active)
{
    cJSON *jsonBuf;

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "platform", "'platform' not found in config file!");
    if (!jsonBuf)
        exit(1);

    cJSON *jsonBufArr = cJSON_CreateArray();

    jsonBuf = jsonBuf->child;
    while (jsonBuf) {
        jsonBufTemp = getJsonObject(jsonBuf, "active", "'active' not found in 'platform' in config file!");
        if (jsonBufTemp) {
            cJSON_DeleteItemFromObject(jsonBuf, "active");
            if ((active == 0) || (jsonBufTemp->valueint == 1)) // Return all or active platforms
                cJSON_AddItemReferenceToArray(jsonBufArr, jsonBuf);
        }

        jsonBuf = jsonBuf->next;
    }

    cJSON_Delete(jsonBuf);

    return jsonBufArr;
}

void setUrlApiVer(void)
{
    cJSON *jsonBuf;
    char strBuf[MAX_URL_LEN];

    jsonBuf = getJsonFile();
    jsonBuf = getJsonObject(jsonBuf, "server", "'server' not found in config file!");
    if (!jsonBuf)
        exit(1);

    jsonBufTemp = getJsonObject(jsonBuf, "url_api", "'url_api' not found in 'server' in config file!");
    if (!jsonBufTemp)
        exit(1);
    strcpy(strBuf, jsonBufTemp->valuestring);

    // TODO: Validate url
    if (strlen(strBuf) < 1) {
        fprintf(stderr, "Server URL is not valid: %s\n", strBuf);
        exit(1);
    }

    strcat(strBuf, "/");
    jsonBufTemp = getJsonObject(jsonBuf, "version", "'version' not found in 'server' in config file!");
    if (!jsonBufTemp)
        exit(1);
    strcat(strBuf, jsonBufTemp->valuestring);
    strcat(strBuf, "/");

    urlApiVer = (char *) malloc(strlen(strBuf) + 1);
    strcpy(urlApiVer, strBuf);

    jsonBufTemp = getJsonObject(jsonBuf, "ssl_verify", "'ssl_verify' not found in 'server' in config file!");
    if (jsonBufTemp)
        sslVerify = jsonBufTemp->valueint;
    else
        sslVerify = 0;

    cJSON_Delete(jsonBuf);
}

void chkServerDependentConfigs(void)
{
    chkPlatform();

    char filePath[PATH_MAX + 1];
    strcpy(filePath, currentPath);
    strcat(filePath, CONFIG_PATH);
    strcat(filePath, ALGO_CRACKER_DIR);

    if (!dirExists(filePath)) {
        mkdirRecursive(filePath);
        cJSON *jsonBuf = getPlatform(0), *reqData = cJSON_CreateArray();
        jsonBuf = jsonBuf->child;
        while (jsonBuf) {
            jsonBufTemp = getJsonObject(jsonBuf, "id", "'id' not found in 'platform' in config file!");
            if (jsonBufTemp)
                cJSON_AddItemReferenceToArray(reqData, jsonBufTemp);

            jsonBuf = jsonBuf->next;
        }
        reqGetAlgoCracker(reqData);

        cJSON_Delete(jsonBuf);
        cJSON_Delete(reqData);
    }
}

const char *getReqUri(int req)
{
    if (req == REQ_GET_VENDOR)
        return URI_GET_VENDOR;
    if (req == REQ_UPDATE_VENDOR)
        return URI_UPDATE_VENDOR;
    if (req == REQ_GET_ALGO_CRACKER)
        return URI_GET_ALGO_CRACKER;
    if (req == REQ_GET_TASK)
        return URI_GET_TASK;
    if (req == REQ_GET_CRACK_INFO)
        return URI_GET_CRACK_INFO;
    if (req == REQ_SEND_RESULT)
        return URI_SEND_RESULT;

    return "";
}

const char *getCracker(const char *mapFilePath, const char *algoId)
{
    if (fileExists(mapFilePath)) {
        char *strBuf;
        if (fileGetContents(&strBuf, mapFilePath, NULL) < 0)
            return "\0";

        cJSON *jsonBuf;
        jsonBuf = cJSON_Parse(strBuf);
        free(strBuf);

        if (jsonBuf) {
            cJSON *jsonMap = jsonBuf->child;
            while (jsonMap) {
                jsonBufTemp = getJsonObject(jsonMap, "algo_id", "'algo_id' not found in algo-cracker file!");
                if (jsonBufTemp) {
                    if (strncmp(algoId, jsonBufTemp->valuestring, strlen(algoId)) == 0) {
                        jsonBufTemp = getJsonObject(jsonMap, "cracker", "'cracker' not found in algo-cracker file!");
                        if (jsonBufTemp)
                            return jsonBufTemp->valuestring;
                    }
                }

                jsonMap = jsonMap->next;
            }
        } else {
            fprintf(stderr, "Invalid JSON in algo-cracker file: %s\n", mapFilePath);
        }
    }

    return "\0";
}

/**
 * cJSON taskInfo {"crack_id":"", "start":"", "offset":"", "platform":""}
 */
int doCrack(const char *crackInfoPath, cJSON **taskInfo)
{
    char *strBuf;
    int ret = -2;

    if (fileGetContents(&strBuf, crackInfoPath, NULL) > -1) {
        cJSON *crackInfo;
        crackInfo = cJSON_Parse(strBuf);
        free(strBuf);

        // cJSON crackInfo {"id":"","gen_name":"","algo_id":"","algo_name":"","lenMin":"","lenMax":"","charset1":"","charset2":"","charset3":"","charset4":"","mask":""}
        if (crackInfo) {
            do { // An once loop to utilize break!
                char pathBuf[PATH_MAX + 1], platformId[20];
                jsonBufTemp = getJsonObject(*taskInfo, "platform", "'platform' not found in response of get task!");
                if (!jsonBufTemp)
                    break;
                strncpy(platformId, jsonBufTemp->valuestring, 19);
                platformId[20] = '\0';
                cJSON_DeleteItemFromObject(*taskInfo, "platform"); // Delete platform, preparing taskInfo to send back to server

                /* Determine crakcer */
                strcpy(pathBuf, currentPath);
                strcat(pathBuf, CONFIG_PATH);
                strcat(pathBuf, ALGO_CRACKER_DIR);
                strcat(pathBuf, PATH_SEPARATOR);
                strcat(pathBuf, platformId);
                strcat(pathBuf, ".force");

                char crackerName[PATH_MAX + 1];
                crackerName[0] = '\0';

                /* Check user defined algo-cracker */
                jsonBufTemp = getJsonObject(crackInfo, "algo_id", "'algo_id' not found in crack info file!");
                if (!jsonBufTemp)
                    break;
                strcpy(crackerName, getCracker(pathBuf, jsonBufTemp->valuestring));

                /* Check default algo-cracker */
                if (strlen(crackerName) < 1) { // Cracker not determined
                    pathBuf[strlen(pathBuf) - 6] = '\0'; // Remove .force
                    jsonBufTemp = getJsonObject(crackInfo, "algo_id", "'algo_id' not found in crack info file!");
                    if (!jsonBufTemp)
                        break;
                    strcpy(crackerName, getCracker(pathBuf, jsonBufTemp->valuestring));
                }

                /* No cracker name found, use default cracker (hashcat) */
                if (strlen(crackerName) < 1) { // Cracker not determined
                    if (strncmp("cpu", platformId, 3)) // Default cpu cracker
                        strcpy(crackerName, "hashcat");
                    else if (strncmp("gpu", platformId, 3)) // Default cpu cracker
                        strcpy(crackerName, "hashcat");
                }

                if (strlen(crackerName) > 0) { // Cracker determined
                    // Check vendor file
                    strcpy(pathBuf, currentPath);
                    strcat(pathBuf, "vendor");
                    strcat(pathBuf, PATH_SEPARATOR);
                    strcat(pathBuf, "cracker");
                    strcat(pathBuf, PATH_SEPARATOR);
                    strcat(pathBuf, crackerName);
                    strcat(pathBuf, PATH_SEPARATOR);
                    strcat(pathBuf, platformId);
                    strcat(pathBuf, PATH_SEPARATOR);
                    strcat(pathBuf, platformId);

                    cJSON *cracker;
                    if (!fileExists(pathBuf)) {
                        cracker = cJSON_CreateObject();
                        cJSON_AddItemToObject(cracker, "object_type", cJSON_CreateString("file"));
                        cJSON_AddItemToObject(cracker, "vendor_type", cJSON_CreateString("cracker"));
                        cJSON_AddItemToObject(cracker, "name", cJSON_CreateString(crackerName));
                        cJSON_AddItemToObject(cracker, "platform_id", cJSON_CreateString(platformId));

                        reqGetVendor(cracker);
                    }

                    char crackerCmd[PATH_MAX + 1];
                    strcpy(crackerCmd, pathBuf);
                    strcat(crackerCmd, " ");

                    // Check vendor info
                    strcpy(pathBuf, getDirName(pathBuf));
                    strcat(pathBuf, PATH_SEPARATOR);
                    strcat(pathBuf, "config.json");
                    if (!fileExists(pathBuf)) {
                        if (cracker) {
                            strcpy(getJsonObject(cracker, "object_type", NULL)->valuestring, "info");
                        } else {
                            cracker = cJSON_CreateObject();
                            cJSON_AddItemToObject(cracker, "object_type", cJSON_CreateString("info"));
                            cJSON_AddItemToObject(cracker, "vendor_type", cJSON_CreateString("cracker"));
                            cJSON_AddItemToObject(cracker, "name", cJSON_CreateString(crackerName));
                            cJSON_AddItemToObject(cracker, "platform_id", cJSON_CreateString(platformId));
                        }

                        reqGetVendor(cracker);
                    }

                    // Prepare crack for execution
                    if (fileGetContents(&strBuf, pathBuf, "Crack info file not valid!") > 0) {
                        cracker = cJSON_Parse(strBuf);
                        free(strBuf);
                        cJSON *jsonBuf = getJsonObject(cracker, "config", "'config' not found in cracker info!");

                        if (jsonBuf) {
                            char char1[255], char2[255], char3[255], char4[255];

                            jsonBuf = getJsonObject(jsonBuf, "args_opt", "'args_opt' not found in cracker info!");
                            if (jsonBuf) {
                                // Prepare CHAR1
                                jsonBufTemp = getJsonObject(jsonBuf, "CHAR1", NULL);
                                if (jsonBufTemp) {
                                    strcpy(char1, jsonBufTemp->valuestring);

                                    jsonBufTemp = getJsonObject(crackInfo, "charset1", NULL);
                                    if (jsonBufTemp && strlen(jsonBufTemp->valuestring) > 0) {
                                        strcpy(char1, strReplace(char1, "CHAR1", jsonBufTemp->valuestring));
                                    } else {
                                        char1[0] = '\0';
                                    }
                                } else {
                                    char1[0] = '\0';
                                }

                                // Prepare CHAR2
                                jsonBufTemp = getJsonObject(jsonBuf, "CHAR2", NULL);
                                if (jsonBufTemp) {
                                    strcpy(char2, jsonBufTemp->valuestring);

                                    jsonBufTemp = getJsonObject(crackInfo, "charset2", NULL);
                                    if (jsonBufTemp && strlen(jsonBufTemp->valuestring) > 0) {
                                        strcpy(char2, strReplace(char2, "CHAR2", jsonBufTemp->valuestring));
                                    } else {
                                        char2[0] = '\0';
                                    }
                                } else {
                                    char2[0] = '\0';
                                }

                                // Prepare CHAR3
                                jsonBufTemp = getJsonObject(jsonBuf, "CHAR3", NULL);
                                if (jsonBufTemp) {
                                    strcpy(char3, jsonBufTemp->valuestring);

                                    jsonBufTemp = getJsonObject(crackInfo, "charset3", NULL);
                                    if (jsonBufTemp && strlen(jsonBufTemp->valuestring) > 0) {
                                        strcpy(char3, strReplace(char3, "CHAR3", jsonBufTemp->valuestring));
                                    } else {
                                        char3[0] = '\0';
                                    }
                                } else {
                                    char3[0] = '\0';
                                }

                                // Prepare CHAR4
                                jsonBufTemp = getJsonObject(jsonBuf, "CHAR4", NULL);
                                if (jsonBufTemp) {
                                    strcpy(char4, jsonBufTemp->valuestring);

                                    jsonBufTemp = getJsonObject(crackInfo, "charset4", NULL);
                                    if (jsonBufTemp && strlen(jsonBufTemp->valuestring) > 0) {
                                        strcpy(char4, strReplace(char4, "CHAR4", jsonBufTemp->valuestring));
                                    } else {
                                        char4[0] = '\0';
                                    }
                                } else {
                                    char4[0] = '\0';
                                }
                            } else {
                                char1[0] = '\0';
                                char2[0] = '\0';
                                char3[0] = '\0';
                                char4[0] = '\0';
                            }

                            /* Prepare cracker command */
                            // Create hashfile
                            jsonBufTemp = getJsonObject(crackInfo, "target", "'target' not found in crack info!");
                            if (!jsonBufTemp)
                                break;
                            strcpy(pathBuf, crackInfoPath);
                            strcpy(pathBuf, getDirName(pathBuf));
                            strcat(pathBuf, PATH_SEPARATOR);
                            strcat(pathBuf, "hashfile");
                            FILE *hashFile;
                            hashFile = fopen(pathBuf, "wb");
                            if (!hashFile) {
                                fprintf(stderr, "Can not create hash file!\n");
                                break;
                            }
                            fputs(jsonBufTemp->valuestring, hashFile);
                            fclose(hashFile);

                            jsonBufTemp = getJsonObject(cracker, "config", "'config' not found in cracker info!");
                            if (jsonBufTemp) {
                                jsonBufTemp = getJsonObject(jsonBufTemp, "args", "'args' not found in cracker info!");
                                if (jsonBufTemp) {
                                    strcat(crackerCmd, jsonBufTemp->valuestring);

                                    /* Replace args with their value */

                                    // Replace START
                                    jsonBufTemp = getJsonObject(*taskInfo, "start", "'start' not found in response of get task!");
                                    if (!jsonBufTemp)
                                        break;
                                    strcpy(crackerCmd, strReplace(crackerCmd, "START", jsonBufTemp->valuestring));

                                    // Replace OFFSET
                                    jsonBufTemp = getJsonObject(*taskInfo, "offset", "'offset' not found in response of get task!");
                                    if (!jsonBufTemp)
                                        break;
                                    strcpy(crackerCmd, strReplace(crackerCmd, "OFFSET", jsonBufTemp->valuestring));

                                    int algoFlag = 0;

                                    // Replace ALGO_ID
                                    jsonBufTemp = getJsonObject(crackInfo, "algo_id", "'algo_id' not found in crack info file!");
                                    if (jsonBufTemp) {
                                        strcpy(crackerCmd, strReplace(crackerCmd, "ALGO_ID", jsonBufTemp->valuestring));
                                        algoFlag = 1;
                                    }

                                    // Replace ALGO_NAME
                                    jsonBufTemp = getJsonObject(crackInfo, "algo_name", "'algo_name' not found in crack info!");
                                    if (jsonBufTemp) {
                                        strcpy(crackerCmd, strReplace(crackerCmd, "ALGO_NAME", jsonBufTemp->valuestring));
                                        algoFlag = 1;
                                    }

                                    if (algoFlag == 0)
                                        break;

                                    // Replace LEN_MIN
                                    jsonBufTemp = getJsonObject(crackInfo, "lenMin", "'lenMin' not found in crack info!");
                                    if (!jsonBufTemp)
                                        break;
                                    strcpy(crackerCmd, strReplace(crackerCmd, "LEN_MIN", jsonBufTemp->valuestring));

                                    // Replace LEN_MAX
                                    jsonBufTemp = getJsonObject(crackInfo, "lenMax", "'lenMax' not found in crack info!");
                                    if (!jsonBufTemp)
                                        break;
                                    strcpy(crackerCmd, strReplace(crackerCmd, "LEN_MAX", jsonBufTemp->valuestring));

                                    // Replace MASK
                                    jsonBufTemp = getJsonObject(crackInfo, "mask", "'mask' not found in crack info!");
                                    if (!jsonBufTemp)
                                        break;
                                    strcpy(crackerCmd, strReplace(crackerCmd, "MASK", jsonBufTemp->valuestring));

                                    // Replace CHAR1
                                    strcpy(crackerCmd, strReplace(crackerCmd, "CHAR1", char1));

                                    // Replace CHAR2
                                    strcpy(crackerCmd, strReplace(crackerCmd, "CHAR2", char2));

                                    // Replace CHAR3
                                    strcpy(crackerCmd, strReplace(crackerCmd, "CHAR3", char3));

                                    // Replace CHAR4
                                    strcpy(crackerCmd, strReplace(crackerCmd, "CHAR4", char4));

                                    // Replace HASH_FILE
                                    strcpy(crackerCmd, strReplace(crackerCmd, "HASH_FILE", pathBuf));

                                    // Replace OUT_FILE
                                    strcat(pathBuf, ".out");
                                    strcpy(crackerCmd, strReplace(crackerCmd, "OUT_FILE", pathBuf));
                                }
                            }

                            /* Execute crack */
                            ret = system(crackerCmd); // Return: <0: not executed successfully; 0: executed successfully; >0: executed with error;

                            if (fileGetContents(&strBuf, pathBuf, NULL) > 0) { // Out file is not empty, so send it to server
                                cJSON_AddStringToObject(*taskInfo, "result", strBuf);
                                free(strBuf);
                            }
                            /* Execution finished */
                        }
                    }
                    cJSON_Delete(cracker);
                } else { // Cracker not determined
                    jsonBufTemp = getJsonObject(*taskInfo, "crack_id", "'crack_id' not found in response of get task!");
                    if (jsonBufTemp)
                        fprintf(stderr, "Can't find a cracker to do the crack. crack_id: %s, platform: %s\n", jsonBufTemp->valuestring, platformId);
                    else
                        fprintf(stderr, "Can't find a cracker to do the crack. crack_id: ?, platform: %s\n", platformId);
                }
            } while (0); // End of once loop

            cJSON_Delete(crackInfo);
        } else {
            fprintf(stderr, "Invalid JSON in crack info file: %s\n", crackInfoPath);
        }
    } else {
        fprintf(stderr, "Can't open crack info file: %s\n", crackInfoPath);
    }

    return ret;
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
        char urlStr[MAX_URL_LEN];

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

            fprintf(stderr, "Can't write response body file!\n");
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
                case REQ_GET_ALGO_CRACKER:
                    resGetAlgoCracker(resBodyFilePath, data);
                    break;
                case REQ_GET_TASK:
                    resGetTask(resBodyFilePath);
                    break;
                case REQ_SEND_RESULT:
                    resSendResult(resBodyFilePath, data);
                    break;
                case REQ_GET_CRACK_INFO:
                    resGetCrackInfo(resBodyFilePath, data);
                    break;
            }
        } else {
            fprintf(stderr, "Server response is not valid. Server url is: %s\n", urlStr);
        }

        // Delete temporary response file
        remove(resBodyFilePath);

        return resCode;
    } else {
        return -1;
    }
}

void reqGetAlgoCracker(cJSON *reqData)
{
    sendRequest(REQ_GET_ALGO_CRACKER, reqData);
}

void resGetAlgoCracker(const char *resBodyPath, cJSON *reqData)
{
    char algoCrackerDirPath[PATH_MAX + 1], algoCrackerPath[PATH_MAX + 1];
    cJSON *jsonPlat;

    strcpy(algoCrackerDirPath, currentPath);
    strcat(algoCrackerDirPath, CONFIG_PATH);
    strcat(algoCrackerDirPath, ALGO_CRACKER_DIR);

    jsonPlat = reqData->child;
    while (jsonPlat) {
        strcpy(algoCrackerPath, algoCrackerDirPath);
        strcat(algoCrackerPath, PATH_SEPARATOR);
        strcat(algoCrackerPath, jsonPlat->valuestring);
        fileCopy(resBodyPath, algoCrackerPath);
        chmod(algoCrackerPath, S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH); // rwx rwx r-x (775)

        jsonPlat = jsonPlat->next;
    }
}

void reqGetVendor(cJSON *vendorData)
{
    sendRequest(REQ_GET_VENDOR, vendorData);
}

void resGetVendor(const char *resBodyPath, cJSON *reqData)
{
    char *strBuf;
    long int resSize;

    resSize = fileGetContents(&strBuf, resBodyPath, "Can't open response file of vendor!");
    if (resSize < 0) {
        return;
    } else {
        if (strncmp(strBuf, "0", resSize) == 0) {
            fprintf(stderr, "Vendor not found in server. object_type: %s, vendor_type: %s, name: %s\n", getJsonObject(reqData, "object_type", NULL)->valuestring,
                getJsonObject(reqData, "vendor_type", NULL)->valuestring, getJsonObject(reqData, "name", NULL)->valuestring);
            return;
        }
    }

    char vendorResPath[PATH_MAX + 1];

    strcpy(vendorResPath, currentPath);
    strcat(vendorResPath, "vendor");
    strcat(vendorResPath, PATH_SEPARATOR);
    strcat(vendorResPath, getJsonObject(reqData, "vendor_type", NULL)->valuestring);
    strcat(vendorResPath, PATH_SEPARATOR);
    strcat(vendorResPath, getJsonObject(reqData, "name", NULL)->valuestring);
    strcat(vendorResPath, PATH_SEPARATOR);
    strcat(vendorResPath, getJsonObject(reqData, "platform_id", NULL)->valuestring);
    strcat(vendorResPath, PATH_SEPARATOR);

    if (strcmp(getJsonObject(reqData, "object_type", NULL)->valuestring, "info") == 0) // Vendor info
        strcat(vendorResPath, "config.json");
    else
        // Vendor file
        strcat(vendorResPath, getJsonObject(reqData, "platform_id", NULL)->valuestring);

    fileCopy(resBodyPath, vendorResPath);

    chmod(vendorResPath, S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH); // rwx rwx r-x (775)
}

void reqUpdateVendor(void)
{
}

void resUpdateVendor(const char *resBodyPath)
{
}

void reqGetTask(void)
{
    cJSON *jsonReqData = cJSON_CreateObject(), *jsonPlatform = cJSON_CreateObject();

    cJSON_AddItemToObject(jsonPlatform, "platform", getPlatform(1));
    cJSON_AddItemToObject(jsonReqData, "client_info", jsonPlatform);

    sendRequest(REQ_GET_TASK, jsonReqData);
    cJSON_Delete(jsonReqData);
}

void resGetTask(const char *resBodyPath)
{
    char *strBuf;

    if (fileGetContents(&strBuf, resBodyPath, "Error in reading response file of get task!") > -1) {
        cJSON *jsonTasks = cJSON_Parse(strBuf);
        free(strBuf);

        if (jsonTasks) {
            cJSON *jsonTask = jsonTasks->child, *jsonResults = cJSON_CreateArray();
            char crackInfoPath[PATH_MAX + 1];
            // TODO:Do cracks of each platform in parallel
            while (jsonTask) {
                // {"crack_id":"","start":"","offset":"","platform":""}
                jsonBufTemp = getJsonObject(jsonTask, "crack_id", "'crack_id' not found in response of get task!");
                if (jsonBufTemp) {
                    strcpy(crackInfoPath, currentPath);
                    strcat(crackInfoPath, "crack");
                    strcat(crackInfoPath, PATH_SEPARATOR);
                    strcat(crackInfoPath, jsonBufTemp->valuestring);
                    strcat(crackInfoPath, PATH_SEPARATOR);
                    strcat(crackInfoPath, "info.json");

                    if (!fileExists(crackInfoPath)) { // Get crack info if needed
                        reqGetCrackInfo(jsonBufTemp->valuestring);
                    }

                    if (fileExists(crackInfoPath)) {
                        cJSON_AddNumberToObject(jsonTask, "status", doCrack(crackInfoPath, &jsonTask));

                        cJSON_AddItemReferenceToArray(jsonResults, jsonTask);
                    } else {
                        fprintf(stderr, "Crack info is not available: %s\n", crackInfoPath);
                    }
                }

                jsonTask = jsonTask->next;
            }

            reqSendResult(jsonResults);

            cJSON_Delete(jsonResults);
            cJSON_Delete(jsonTasks);
        } else {
            fprintf(stderr, "Invalid JSON response file for get task!\n");
        }
    }
}

void reqSendResult(cJSON *jsonResults)
{
    sendRequest(REQ_SEND_RESULT, jsonResults);
}

void resSendResult(const char *resBodyPath, cJSON *reqData)
{
    /* Delete the recieved result files */
    char outFilePath[PATH_MAX + 1];
    cJSON *result = reqData->child;
    while (result) {
        jsonBufTemp = getJsonObject(result, "crack_id", "'crack_id' not found in response of send result!");
        if (jsonBufTemp) {
            strcpy(outFilePath, currentPath);
            strcat(outFilePath, "crack");
            strcat(outFilePath, PATH_SEPARATOR);
            strcat(outFilePath, jsonBufTemp->valuestring);
            strcat(outFilePath, PATH_SEPARATOR);
            strcat(outFilePath, "hashfile.out");
            if (fileExists(outFilePath))
                unlink(outFilePath);
        }

        result = result->next;
    }
}

void reqGetCrackInfo(const char *crackId)
{
    cJSON *jsonReqData = cJSON_CreateObject();

    cJSON_AddItemToObject(jsonReqData, "id", cJSON_CreateString(crackId));

    sendRequest(REQ_GET_CRACK_INFO, jsonReqData);
    cJSON_Delete(jsonReqData);
}

void resGetCrackInfo(const char *resBodyPath, cJSON *reqData)
{
    char crackInfoPath[PATH_MAX + 1];

    strcpy(crackInfoPath, currentPath);
    strcat(crackInfoPath, "crack");
    strcat(crackInfoPath, PATH_SEPARATOR);
    strcat(crackInfoPath, getJsonObject(reqData, "id", NULL)->valuestring);
    strcat(crackInfoPath, PATH_SEPARATOR);
    strcat(crackInfoPath, "info.json");

    fileCopy(resBodyPath, crackInfoPath);

    chmod(crackInfoPath, S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH); // rwx rwx r-x (775)
}
