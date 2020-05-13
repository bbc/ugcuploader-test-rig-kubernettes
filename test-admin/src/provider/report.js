import fetch from "isomorphic-fetch";

const fetchPostsApi = () => {
  let data = fetch("/tenants").then(function (response) {
    return response.json();
  });

  return data;
};

const fetchReportByTenant = (tenantId) => {
  let data = fetch("/tenantReport?tenant=" + tenantId).then(function (
    response
  ) {
    return response.json();
  });

  return data;
};

const fetchDashboardUrl = () => {
  let data = fetch("/dashboardUrl").then(function (response) {
    if (response.status == 500) {
      return {
        DashboardURL: "",
        ChronografURL: "",
        ReportURL: "",
        MonitorURL: "",
      };
    } else {
      return response.json();
    }
  });

  return data;
};

const fetchSlavesForTenant = (tenantId) => {
  let data = fetch("/slaves?tenant=" + tenantId).then(function (response) {
    return response.json();
  });

  return data;
};

export {
  fetchPostsApi,
  fetchReportByTenant,
  fetchDashboardUrl,
  fetchSlavesForTenant,
};
