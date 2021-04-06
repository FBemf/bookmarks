let partials = (() => {
    const SPINNER_TIMEOUT = 500

    const SPINNER_HTML = `<div class="spinner">
    <div class="bounce1"></div>
    <div class="bounce2"></div>
    <div class="bounce3"></div>
    </div>`

    const ERROR_HTML = "Error: Retry operation"

    function replace(target, html) {
        for (element of document.getElementsByClassName(target)) {
            element.innerHTML = html
        }
    }

    function get(endpoint, target) {
        let spinnerTimeout = setTimeout(() => {
            for (element of document.getElementsByClassName(target)) {
                element.innerHTML = SPINNER_HTML
            }
        }, SPINNER_TIMEOUT)
        fetch(`/partial/${endpoint}`).then(
            (response) => {
                if (response.ok) {
                    clearTimeout(spinnerTimeout)
                    response.text().then((text) => {
                        replace(target, text)
                    })
                } else {
                    clearTimeout(spinnerTimeout)
                    replace(target, ERROR_HTML)
                    console.log(`partial request to ${endpoint} failed with code ${response.status}`)
                }
            }
        )
    }

    return {
        get: get
    }
})()