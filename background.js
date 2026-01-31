function log(msg) {
    const now = new Date().toISOString();
    console.log(`[${now}] ${msg}`);
}

const port = chrome.runtime.connectNative("rofi_chrome_tab");

port.onMessage.addListener((msg) => {
    log('onMessage: ' + JSON.stringify(msg));

    if (msg.command === 'list') {
        chrome.tabs.query({})
            .then(tabs => {
                let tabs1 = [];
                for (let tab of tabs) {
                    const host = (() => {
                        try {
                            return new URL(tab.url).hostname;
                        } catch {
                            return 'No URL'
                        }
                    })();
                    tabs1.push({ id: tab.id, title: tab.title, host });
                }

                const message = JSON.stringify(tabs1);
                const preview = message.length >= 30 ? message.slice(0, 30) + "..." : message;
                log('postMessage: ' + preview);
                port.postMessage(tabs1);
            });
        return;
    }

    if (msg.command === 'select') {
        chrome.tabs.update(msg.tabId, { active: true })
            .then(tab => {
                chrome.windows.update(tab.windowId, { focused: true })
            });
        return;
    }

    if (msg.command === 'count') {
        chrome.tabs.query({})
            .then(tabs => {
                log('postMessage: ' + tabs.length);
                port.postMessage(tabs.length);
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

function notifyUpdatedEvent() {
    chrome.tabs.query({})
        .then(tabs => {
            let tabs1 = [];
            for (let tab of tabs) {
                const host = (() => {
                    try {
                        return new URL(tab.url).hostname;
                    } catch {
                        return 'No URL'
                    }
                })();
                tabs1.push({ id: tab.id, title: tab.title, host });
            }

            const message = JSON.stringify(tabs1);
            const preview = message.length >= 30 ? message.slice(0, 30) + "..." : message;
            log('postMessage: ' + preview);
            port.postMessage({
                'type': 'updated',
                'tabs': tabs1
            });
    });
}

chrome.tabs.onActivated.addListener((_tabId, _moveInfo) => {
    notifyUpdatedEvent();
});

notifyUpdatedEvent();
