// Constants
const PREVIEW_LENGTH = 30;
const DEFAULT_HOST = 'No URL';

/**
 * Logs a message with timestamp
 * @param {string} msg - Message to log
 */
function log(msg) {
    const now = new Date().toISOString();
    console.log(`[${now}] ${msg}`);
}

/**
 * Extracts hostname from a URL
 * @param {string} url - The URL to extract hostname from
 * @returns {string} The hostname or default message
 */
function getHostFromUrl(url) {
    try {
        return new URL(url).hostname;
    } catch {
        return DEFAULT_HOST;
    }
}

/**
 * Processes tabs into a simplified format
 * @param {Array} tabs - Array of Chrome tab objects
 * @returns {Array} Processed tabs with id, title, and host
 */
function processTabs(tabs) {
    return tabs.map(tab => ({
        id: tab.id,
        title: tab.title,
        host: getHostFromUrl(tab.url)
    }));
}

/**
 * Creates a preview of a message
 * @param {string} message - The message to preview
 * @returns {string} Preview string
 */
function createPreview(message) {
    return message.length >= PREVIEW_LENGTH 
        ? message.slice(0, PREVIEW_LENGTH) + "..." 
        : message;
}

/**
 * Logs a preview of processed tabs
 * @param {Array} processedTabs - The processed tabs to log
 */
function logTabsPreview(processedTabs) {
    const message = JSON.stringify(processedTabs);
    const preview = createPreview(message);
    log('postMessage: ' + preview);
}

const port = chrome.runtime.connectNative("rofi_chrome_tab");

port.onMessage.addListener((msg) => {
    log('onMessage: ' + JSON.stringify(msg));

    if (msg.command === 'list') {
        chrome.tabs.query({})
            .then(tabs => {
                const processedTabs = processTabs(tabs);
                logTabsPreview(processedTabs);
                port.postMessage(processedTabs);
            })
            .catch(error => {
                console.error('Error listing tabs:', error);
            });
        return;
    }

    if (msg.command === 'select') {
        chrome.tabs.update(msg.tabId, { active: true })
            .then(tab => {
                return chrome.windows.update(tab.windowId, { focused: true });
            })
            .catch(error => {
                console.error('Error selecting tab:', error);
            });
        return;
    }

    if (msg.command === 'count') {
        chrome.tabs.query({})
            .then(tabs => {
                log('postMessage: ' + tabs.length);
                port.postMessage(tabs.length);
            })
            .catch(error => {
                console.error('Error counting tabs:', error);
            });
        return;
    }
    
    console.log('Invalid command: ' + msg.command);
});

port.onDisconnect.addListener(() => {
    if (chrome.runtime.lastError) {
        console.error(chrome.runtime.lastError.message);
    }
});

/**
 * Notifies about tab updates
 */
function notifyUpdatedEvent() {
    chrome.tabs.query({})
        .then(tabs => {
            const processedTabs = processTabs(tabs);
            logTabsPreview(processedTabs);
            port.postMessage({
                type: 'updated',
                tabs: processedTabs
            });
        })
        .catch(error => {
            console.error('Error notifying update:', error);
        });
}

chrome.tabs.onActivated.addListener(() => {
    notifyUpdatedEvent();
});

notifyUpdatedEvent();
