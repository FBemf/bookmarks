(() => {
    const application = Stimulus.Application.start()

    application.register("are-you-sure", class extends Stimulus.Controller {
        static get targets() {
            return ["initial", "primary"]
        }

        prime() {
            this.initialTarget.style.display = "none"
            this.primaryTarget.style.display = ""
        }

        cancel() {
            this.initialTarget.style.display = ""
            this.primaryTarget.style.display = "none"
        }
    })


    application.register("new-dialogue", class extends Stimulus.Controller {
        static get targets() {
            return ["form", "showButton", "hideButton"]
        }

        show() {
            this.formTarget.style.display = ""
            this.showButtonTarget.style.display = "none"
            this.hideButtonTarget.style.display = ""
        }

        hide() {
            this.formTarget.style.display = "none"
            this.showButtonTarget.style.display = ""
            this.hideButtonTarget.style.display = "none"
        }
    })

    application.register("bookmark-tagger", class extends Stimulus.Controller {
        static get targets() {
            return ["tagName", "tagList"]
        }

        addTag(event) {
            this.internalAddTag(event, "tag")
        }

        addSearchTag(event) {
            this.internalAddTag(event, "searchTag")
        }

        internalAddTag(event, fieldName) {
            if (event.type == "keydown") {
                if (event.key == "Enter") {
                    event.preventDefault()
                } else {
                    return
                }
            }
            let name = this.tagNameTarget.value
            if (name != "") {
                this.tagNameTarget.value = ""
                let newTag = document.createElement("div")
                this.tagListTarget.appendChild(newTag)
                newTag.outerHTML = `
                    <span class=taglist__tag data-controller="tag" data-tag-target="self">
                        <input type=hidden name="${fieldName}" readonly=readonly value="${name}">
                            ${name}
                            <button class=link data-action="click->tag#remove" type=button>×</button>
                            &nbsp;
                    </span>`
            }
        }
    });

    application.register("tag", class extends Stimulus.Controller {
        static get targets() {
            return ["self"]
        }

        remove() {
            this.selfTarget.parentNode.removeChild(this.selfTarget)
        }
    })

    application.register("text-copier", class extends Stimulus.Controller {
        static get targets() {
            return ["text"]
        }

        copy() {
            let displayValue = this.textTarget.style.display
            this.textTarget.style.display = ""
            this.textTarget.select()
            document.execCommand("copy")
            this.textTarget.style.display = displayValue
        }
    })

    application.register("bookmarklet-copier", class extends Stimulus.Controller {
        static get targets() {
            return ["key"]
        }

        copy() {
            let textElement = document.body.appendChild(document.createElement("textarea"))
            let port = ""
            if (window.location.port != "") {
                port = ":" + window.location.port
            }
            textElement.innerText = `
javascript:(()=>{
let auth = "Bearer ${this.keyTarget.value}";
let name = window.prompt("Name", document.title);
if (name == null) { return; }
let url = window.prompt("URL", window.location.href);
if (url == null) { return; }
let description = window.prompt("Description", "Type some stuff here");
if (description == null) { return; }
let tags = [];
while (true) {
let got = window.prompt("Add tag?", "");
if (got == null || got == "") { break; }
tags.push(got);
}
let queryUrl = "${window.location.protocol}//${window.location.hostname}${port}/api/newbookmark";
try {
fetch(queryUrl, {
method: "POST",
headers: {"Authorization": auth},
body: JSON.stringify({name: name, url: url, description: description, tags: tags}),
})} catch(e) {
alert("Failed to create bookmark!\n" + String(e))
}})()`
            textElement.select()
            document.execCommand("copy")
            textElement.parentNode.removeChild(textElement)
        }
    })
})()