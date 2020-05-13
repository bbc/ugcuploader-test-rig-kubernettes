import React, {Component} from 'react';
import {
  Container, 
  Header, 
  Message, 
  Tab 
} from 'semantic-ui-react';
import Jmeter from './Jmeter/Jmeter';
import WeaveScope from './Weavescope/WeaveScope';
import Report from './Report/Report';
import Graphana from './Grafana/Graphana';
import ChronografContainer from './Chronograf/ChronografContainer';
import './Main.css';


const panes = [
  { menuItem: 'Jmeter', render: () => <Tab.Pane><Jmeter/></Tab.Pane> },
  { menuItem: 'Chronograf', render: () => <Tab.Pane><ChronografContainer/></Tab.Pane> },
  { menuItem: 'Grafana', render: () => <Tab.Pane><Graphana/></Tab.Pane> },
  { menuItem: 'Weavscope', render: () => <Tab.Pane><WeaveScope/></Tab.Pane> },
  { menuItem: 'Report', render: () => <Tab.Pane><Report/></Tab.Pane> },
]

class Main extends Component {
  state = {}
    render() {
      return (
        <Container style={{ width: 'auto'}}className="Main-Wrapper">
          <Container textAlign='center'><Header as="h1">Kubernetes Load Test Rig</Header></Container>
          <Container className="Message-wrapper" textAlign='center'>
          <Message warning>
            <Message.Header textalign="center"> Embedding grafana triggers firefox security!</Message.Header>
            <p>
              https://stackoverflow.com/questions/11768221/firefox-websocket-security-issue/12042843#12042843
                It works fine in chrome.
            </p>
          </Message>
          </Container>
    
          <Tab  menu={{ color: 'blue', attached: true, tabular: true }} panes={panes} />
        </Container>
      );
    }
  }

export default Main;