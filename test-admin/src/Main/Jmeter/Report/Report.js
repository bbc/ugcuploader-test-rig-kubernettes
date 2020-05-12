import React, {Component} from 'react';
import {
  Container, 
  Header, 
  Tab 
} from 'semantic-ui-react';
import GenerateGraph from './GenerateGraph/GenerateGraph';
import SlaveLogs from './SlaveLogFiles/SlaveLogs';
import './Report.css';


const panes = [
  { menuItem: 'Generate Graphs', render: () => <Tab.Pane><GenerateGraph/></Tab.Pane> },
  { menuItem: 'JMeter Log Files', render: () => <Tab.Pane><SlaveLogs/></Tab.Pane> }
]

class JmeterTestReports extends Component {
  state = {}
    render() {
      return (
        <Container className="Main-Wrapper">
          <Container textAlign='center'><Header as="h1">Jmeter Test Reports</Header></Container>
          <Tab  menu={{ color: 'blue', attached: true, tabular: true }} panes={panes} />
        </Container>
      );
    }
  }

export default JmeterTestReports;