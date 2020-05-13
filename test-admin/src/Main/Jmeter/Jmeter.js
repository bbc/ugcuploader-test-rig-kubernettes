import React, { Component } from "react";
import { Container, Grid } from "semantic-ui-react";
import CreateTest from './Create/Create';
import JmeterTestReports from './Report/Report';
import Monitor from './Monitor/Monitor';
import "./Jmeter.css";

class Jmeter extends Component {
  state = {};
  render() {
    return (
      <Container style={{width: 'fit-content'}} className="Jmeter-Wrapper">
        <Grid divided>
          <Grid.Row columns={3}>
            <Grid.Column>
            <JmeterTestReports/>
            </Grid.Column>
            <Grid.Column>
             <CreateTest/>
            </Grid.Column>
            <Grid.Column>
             <Monitor/>
            </Grid.Column>
          </Grid.Row>

        </Grid>
      </Container>
    );
  }
}

export default Jmeter;
