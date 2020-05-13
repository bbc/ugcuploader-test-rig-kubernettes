import React, { Component } from "react";
import { Grid } from "semantic-ui-react";
import StopTest from "./StopTest/StopTest";
import TestStatus from "./TestStatus/TestStatus";
import NodeInformation from "./NodeInformation/NodeInformation";
import TenantDeleton from "./TenantDeletion/TenantDeletion";
import "./Monitor.css";

class Monitor extends Component {
  state = {};
  render() {
    return (
      <Grid celled="internally" columns={1}>
        <Grid.Row>
          <Grid.Column>
            <TenantDeleton />
          </Grid.Column>
        </Grid.Row>
        <Grid.Row>
          <Grid.Column>
            <StopTest />
          </Grid.Column>
        </Grid.Row>
        <Grid.Row>
          <Grid.Column>
            <TestStatus />
          </Grid.Column>
        </Grid.Row>
        <Grid.Row>
          <Grid.Column style={{ overflow: "auto" }}>
            <NodeInformation />
          </Grid.Column>
        </Grid.Row>
      </Grid>
    );
  }
}

export default Monitor;
