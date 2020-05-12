import React, { Component } from "react";

import MyConsumer from "../../MyConsumer";
import "./Graphana.css";

export class Graphana extends Component {
  state = {
    items: {}
  };

  render() {
    return (
      <MyConsumer>
        {({ graphanaUrl }) => (
    <object data={ graphanaUrl} width="100%" height="100%" >
    </object>
        )}
      </MyConsumer>
    );
  }
}

export default Graphana;
