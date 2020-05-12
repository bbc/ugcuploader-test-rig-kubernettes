import fetch from 'isomorphic-fetch'

const fetchPostsApi = ()=> {
  let data = fetch('/tenants').then(function(response) {
    console.log(JSON.stringify(response));
    return response.json();
  });
  
  return data;
}

const fetchReportByTenant = (tenantId) => {
  let data = fetch('/tenantReport?tenant='+tenantId).then(function(response) {
    console.log(JSON.stringify(response));
    return response.json();
  });
  
  return data;
}

const fetchDashboardUrl = () => {
  let data = fetch('/dashboardUrl').then(function(response) {
    console.log(JSON.stringify(response));
    return response.json();
  });
  
  return data;
}

export { fetchPostsApi, fetchReportByTenant, fetchDashboardUrl};