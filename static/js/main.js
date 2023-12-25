


let switchTheme = document.getElementById('switch-theme')
let mode = localStorage.getItem('mode') ?? 'light'
ui('mode', mode)
switchTheme.checked = mode == 'dark'

switchTheme.addEventListener('click', (target) => {
    let mode = localStorage.getItem('mode') ?? 'light'
    let newMode = mode == 'light' ? 'dark' : 'light'
    localStorage.setItem('mode', newMode)
    ui('mode', newMode)
})
