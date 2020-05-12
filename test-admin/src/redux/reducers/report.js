import { REPORT_FETCH_TENANTS, RECEIVE_REPORT_TENANTS } from '../actionTypes';
const reportFetchTenants = (state = [], action) => {
    switch (action.type) {
        case REPORT_FETCH_TENANTS:
          return state;
        case RECEIVE_REPORT_TENANTS:
          return action.tenants;
        default:
        return state;
    }
  }
  
  export default reportFetchTenants;