function getID() {
    let id = localStorage.getItem('id')
    if(id == null) {
        fetch('http://localhost:8080/v1/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                machineName: "desktop",
                trackerName: "firefox"
            })
        }).then(response => {
            return response.json()
        }).then(data => {
            localStorage.setItem('id', data.id)
        }).catch(err => {
            console.log(err)
        })
    }

    return id;
}

browser.tabs.onActivated.addListener((activeInfo) => {
    let id = getID();
    console.log(browser.tabs.get(activeInfo.tabId));
    
    fetch('http://localhost:8080/v1/send_data', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            id: id,
            label: browser.tabs.get(activeInfo.tabId).title,
            state: "ENTER"
        })
    }).catch(err => {
        console.log(err)
    });
});
