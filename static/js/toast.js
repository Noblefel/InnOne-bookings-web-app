function showToast(messages, header, colorClass = 'error') {

    const id = `toast-${Math.ceil(Math.random() * 10000)}`;

    const li = typeof messages == 'string' 
    ? messages
    : messages.map(x => `<li>${x}</li>`).join('')

    let toast = document.createElement('div')
    toast.setAttribute('class',`snackbar ${colorClass}`)
    toast.setAttribute('id', id)
    toast.innerHTML = `
    ${header ? `<p>${header}</p>` : ''}
    <ul>${li}</ul>
    `

    document.body.appendChild(toast);

    ui('#' + id, 7000) 

    setTimeout(() => {
        toast.remove()
    }, 7000);
}  