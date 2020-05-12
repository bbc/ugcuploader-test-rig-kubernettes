import React, { Component } from "react";

import MyConsumer from "../../MyConsumer";
import "./WeaveScope.css";

export class WeaveScope extends Component {
  state = {
    items: {}
  };

  render() {
    return (
      <MyConsumer>
        {({ weaveScopeUrl }) => (
    <object data={ weaveScopeUrl} width="100%" height="100%" >
    </object>
        )}
      </MyConsumer>
    );
  }
}

export default WeaveScope;
