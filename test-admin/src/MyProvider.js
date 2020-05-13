import MyContext from "./MyContext";

import React, { Component } from "react";

import { fetchReportByTenant, fetchDashboardUrl, fetchSlavesForTenant } from './provider/report';
import _ from 'lodash';

class MyProvider extends Component {
  constructor(props) {
    super(props);
    this.state = {
      reports: [],
      slaves: [],
      graphanaUrl: '',
      chronographUrl: '',
      graphsUrl: '',
      weaveScopeUrl: ''
    };
  }

  async componentDidMount() {
    this.updateFunctions();
    this.fetchDashboardURLS();
  }

  updateFunctions = () => {
     this.setState({fetchSlaves: this.fetchSlaves});
  }
  fetchReportsForTenant = async (id) => {
    const report = await fetchReportByTenant(id);
    let reports = _.map(report, (item)=> {
      return {id: id.concat("-",item.date), date: item.date}
    });
    this.setState({ reports: reports });
  };

  fetchReportsForTenant = async (id) => {
    const report = await fetchReportByTenant(id);
    let reports = _.map(report, (item)=> {
      return {id: id.concat("-",item.date), date: item.date}
    });
    this.setState({ reports: reports });
  };

  fetchSlaves = async(tenant) => {
    const slaves = await fetchSlavesForTenant(tenant);
    this.setState({ slaves: slaves });
  }

  fetchDashboardURLS = async () => {
    const dashboard = await fetchDashboardUrl();
    this.setState({ graphanaUrl: dashboard.DashboardURL });
    this.setState({ chronographUrl: dashboard.ChronografURL });
    this.setState({ graphsUrl: dashboard.ReportURL });
    this.setState({ weaveScopeUrl: dashboard.MonitorURL });
  };
  render() {
    return (
      <MyContext.Provider
        value={{
          ...this.state,
        }}
      >
        {this.props.children}
      </MyContext.Provider>
    );
  }
}

export default MyProvider;
