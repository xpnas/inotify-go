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

export function githubLogin(params) {
  return request({ url: '/oauth/githublogin', method: 'get', params })
}

export function githubBind(params) {
  return request({ url: '/oauth/githubbind', method: 'get', params })
}

export function githubUnbind() {
  return request({ url: '/oauth/githubunbind', method: 'post' })
}

export function getWeixinQrEnable() {
  return request({ url: '/oauth/WeixinQrEnable', method: 'get' })
}

export function weixinQrLogin(params) {
  return request({ url: '/oauth/WeixinQrLogin', method: 'get', params })
}

export function weixinQrBind(params) {
  return request({ url: '/oauth/WeixinQrBind', method: 'get', params })
}

export function weixinQrUnbind() {
  return request({ url: '/oauth/WeixinQrUnbind', method: 'post' })
}
