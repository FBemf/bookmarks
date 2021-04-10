(() => {
    const application = Stimulus.Application.start()

    application.register("bookmark-deleter", class extends Stimulus.Controller {
        static get targets() {
            return ["deletePrimary", "deleteYN"]
        }

        delete() {
            this.deletePrimaryTarget.style.display = "none"
            this.deleteYNTarget.style.display = ""
        }

        deleteCancel() {
            this.deletePrimaryTarget.style.display = ""
            this.deleteYNTarget.style.display = "none"
        }
    })


    application.register("new-bookmark", class extends Stimulus.Controller {
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
                let newTag = document.createElement("li")
                this.tagListTarget.appendChild(newTag)
                newTag.outerHTML = `<li data-controller="tag" data-tag-target="self">
                        <input type=text name=searchTag readonly=readonly value="${name}">
                        <button data-action="click->tag#remove" type=button>Remove</button>
                    </li>`
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
})()