import React, {Component} from 'react';
import {
  Container, 
  Header
} from 'semantic-ui-react';
import './NodeInformation.css';

class NodeInformation extends Component {
  state = {}
    render() {
      return (

        <Container className="Main-Wrapper">
        <Container textAlign='center'><Header as="h1">Node Information</Header></Container>
        <div> Information about the nodes</div>
        </Container>
       
      );
    }
  }

export default NodeInformation;