import React, { Component } from "react";

import MyConsumer from "../../MyConsumer";
import { Segment } from "semantic-ui-react";
import "./WeaveScope.css";

export class WeaveScope extends Component {
  state = {
    items: {},
  };

  render() {
    return (
      <MyConsumer>
        {({ weaveScopeUrl }) => (
          <Segment className="Weavescope">
            <object data={weaveScopeUrl} width="100%" height="100%">
                Kubernetes Monitor
            </object>
          </Segment>
        )}
      </MyConsumer>
    );
  }
}

export default WeaveScope;
