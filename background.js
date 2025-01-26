const port = chrome.runtime.connectNative("rofi_chrome_tab");

port.onMessage.addListener((msg) => {
    if (msg.command === 'list') {
        chrome.tabs.query({})
            .then(tabs => {
                let titles = [];
                for (let tab of tabs) {
                    titles.push({ id: tab.id, title: tab.title, host: new URL(tab.url).hostname });
                }
                port.postMessage(titles);
            });
        return;
    }

    if (msg.command === 'select') {
        chrome.tabs.update(msg.tabId, { active: true })
            .then(tab => {
                chrome.windows.update(tab.windowId, { focused: true })
                    .then(window => {
                        port.postMessage('ok');
                    });
            });
        return;
    }

    if (msg.command === 'count') {
        chrome.tabs.query({})
            .then(tabs => {
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
