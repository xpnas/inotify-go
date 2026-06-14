import request from './request'

export function getGlobal() {
  return request({ url: '/settingsys/GetGlobal', method: 'get' })
}

export function setGlobal(params) {
  return request({ url: '/settingsys/SetGlobal', method: 'post', params })
}

export function getJWT() {
  return request({ url: '/settingsys/GetJWT', method: 'get' })
}

export function setJWT(data) {
  return request({ url: '/settingsys/SetJWT', method: 'post', data })
}

export function getUsers() {
  return request({ url: '/settingsys/GetUsers', method: 'get' })
}

export function activeUser(userName, state) {
  return request({ url: '/settingsys/ActiveUser', method: 'post', params: { userName, state } })
}

export function deleteUser(userName) {
  return request({ url: '/settingsys/DeleteUser', method: 'post', params: { userName } })
}

export function getSendInfos() {
  return request({ url: '/settingsys/GetSendInfos', method: 'get' })
}

export function getSendTypeInfos() {
  return request({ url: '/settingsys/GetSendTypeInfos', method: 'get' })
}

export function getDiagnostics() {
  return request({ url: '/settingsys/Diagnostics', method: 'get', silentError: true })
}

export function backupDatabase() {
  return request({ url: '/settingsys/BackupDatabase', method: 'get', responseType: 'blob' })
}
