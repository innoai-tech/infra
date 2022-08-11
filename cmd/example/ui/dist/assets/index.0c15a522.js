(async () => {
    const res = await fetch("/api/example")
    const data = await res.json();
    document.querySelector("#root").innerHTML = JSON.stringify(data, null, 2)
})()
