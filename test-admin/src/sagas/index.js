import { put, call, fork, select } from 'redux-saga/effects'
import fetch from 'isomorphic-fetch'
import * as actions from '../redux/actions'
import { reportFetchTenantsSelector } from '../redux/selectors'

export function fetchPostsApi() {
  let data = fetch('/tenants').then(function(response) {
    return response.json();
  });
  
  //let js = await response.json();
  //console.log(response);
  return data;
}

export function* reportFetchTenants() {
  yield put(actions.reportFetchTenants());
  const tenants = yield call(fetchPostsApi);
  yield put(actions.receiveReportTenants(tenants))
}

export function* startup() {
  const selectedTenants = yield select(reportFetchTenantsSelector)
  yield fork(reportFetchTenants, selectedTenants)
}

export default function* root() {
  yield fork(startup)
}
