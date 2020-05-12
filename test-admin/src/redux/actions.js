import { REPORT_FETCH_TENANTS, RECEIVE_REPORT_TENANTS } from './actionTypes';

export function reportFetchTenants() {
    return {
        type: REPORT_FETCH_TENANTS
    };
}

export function receiveReportTenants(tenants) {
    return {
      type: RECEIVE_REPORT_TENANTS,
      tenants,
      receivedAt: new Date().setMilliseconds(0),
    }
  }