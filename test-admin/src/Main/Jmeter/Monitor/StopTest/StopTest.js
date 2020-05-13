import React, {Component} from 'react';
import {
  Container, 
  Header, 
  Tab 
} from 'semantic-ui-react';
import Safely from './Safely/Safely';
import ForceStop from './ForceStop/ForceStop';
import './StopTest.css';


const panes = [
  { menuItem: 'Safely', render: () => <Tab.Pane><Safely/></Tab.Pane> },
  { menuItem: 'Force', render: () => <Tab.Pane><ForceStop/></Tab.Pane> }
]

class StopTest extends Component {
  state = {}
    render() {
      return (
        <Container className="Main-Wrapper">
          <Container textAlign='center'><Header as="h1">Stop Tests</Header></Container>
          <Tab  menu={{ color: 'blue', attached: true, tabular: true }} panes={panes} />
        </Container>
      );
    }
  }

export default StopTest;