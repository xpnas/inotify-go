import request from './request'

export function getSetting() {
  return request({ url: '/setting', method: 'get' })
}

export function getSendTemplates() {
  return request({ url: '/setting/GetSendTemplates', method: 'get' })
}

export function getSendAuths() {
  return request({ url: '/setting/GetSendAuths', method: 'get' })
}

export function getMessageHistories(params) {
  return request({ url: '/setting/GetMessageHistories', method: 'get', params })
}

export function addSendAuth(data) {
  return request({ url: '/setting/AddSendAuth', method: 'post', data })
}

export function modifySendAuth(data) {
  return request({ url: '/setting/ModifySendAuth', method: 'post', data })
}

export function activeSendAuth(sendAuthId, state) {
  return request({ url: '/setting/ActiveSendAuth', method: 'post', params: { sendAuthId, state } })
}

export function deleteSendAuth(sendAuthId) {
  return request({ url: '/setting/DeleteSendAuth', method: 'post', params: { sendAuthId } })
}

export function reSendKey(sendAuthId) {
  return request({ url: '/setting/reSendKey', method: 'get', params: { sendAuthId } })
}
