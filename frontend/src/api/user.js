import request from './request'

export function login(data) {
  return request({ url: '/oauth/login', method: 'post', data })
}

export function getInfo() {
  return request({ url: '/oauth/info', method: 'get' })
}

export function logout() {
  return request({ url: '/oauth/logout', method: 'post' })
}

export function resetPassword(params) {
  return request({ url: '/oauth/resetPassword', method: 'post', params })
}

export function getGithubEnable() {
  return request({ url: '/oauth/GithubEnable', method: 'get' })
}

export function githubLogin(code) {
  return request({ url: '/oauth/githublogin', method: 'get', params: { code } })
}

export function getWeixinQrEnable() {
  return request({ url: '/oauth/WeixinQrEnable', method: 'get' })
}

export function weixinQrLogin(params) {
  return request({ url: '/oauth/WeixinQrLogin', method: 'get', params })
}
