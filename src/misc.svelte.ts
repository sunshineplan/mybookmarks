import JSEncrypt from 'jsencrypt'
import Swal, { type SweetAlertIcon } from 'sweetalert2'

class Toggler {
  status = $state(false)
  toggle() { this.status = !this.status }
  close() { this.status = false }
}
export const showSidebar = new Toggler

class Loading {
  #n = $state(0)
  show = $derived(this.#n > 0)
  start() { this.#n += 1 }
  end() { this.#n -= 1 }
}
export const loading = new Loading

export const encrypt = (pubkey: string, password: string) => {
  const encrypt = new JSEncrypt()
  encrypt.setPublicKey(pubkey)
  const s = encrypt.encrypt(password)
  if (s === false) return password
  return s
}

export const fire = async (title?: string, html?: string, icon?: SweetAlertIcon) => {
  const swal = Swal.mixin({
    customClass: { confirmButton: 'swal btn btn-primary' },
    buttonsStyling: false
  })
  await swal.fire(title, html, icon)
  if (title == 'Fatal') throw html
}

export const valid = () => {
  let result = true
  Array.from(document.querySelectorAll('input'))
    .forEach(i => { if (!i.checkValidity()) result = false })
  return result
}

export const poll = async (signal: AbortSignal) => {
  let resp: Response
  try {
    resp = await fetch('/poll', { signal })
  } catch (e) {
    let message = ''
    if (typeof e === "string") {
      message = e
    } else if (e instanceof Error) {
      message = e.message
    }
    resp = new Response(message, { "status": 500 })
  }
  return resp
}

export const post = async (url: string, data?: any, universal?: boolean) => {
  let resp: Response
  const init: RequestInit = {
    method: 'post',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  }
  if (universal) init.credentials = 'include'
  loading.start()
  try {
    resp = await fetch(url, init)
  } catch (e) {
    let message = ''
    if (typeof e === "string") {
      message = e
    } else if (e instanceof Error) {
      message = e.message
    }
    resp = new Response(message, { "status": 500 })
  }
  loading.end()
  if (resp.status == 401) {
    await fire('Error', 'Login status has changed. Please Re-login!', 'error')
    window.location.href = '/'
  } else if (resp.status == 409) {
    await fire('Error', 'Data has changed.', 'error')
    window.location.href = '/'
  }
  return resp
}

export const confirm = async (type: string) => {
  const confirm = await Swal.fire({
    title: 'Are you sure?',
    text: 'This ' + type + ' will be deleted permanently.',
    icon: 'warning',
    confirmButtonText: 'Delete',
    showCancelButton: true,
    focusCancel: true,
    customClass: {
      confirmButton: 'swal btn btn-danger',
      cancelButton: 'swal btn btn-primary'
    },
    buttonsStyling: false
  })
  return confirm.isConfirmed
}

export const pasteText = (event: ClipboardEvent) => {
  event.preventDefault()
  if (event.clipboardData) {
    document.execCommand('insertHTML', false, event.clipboardData.getData('text/plain'))
  }
}
