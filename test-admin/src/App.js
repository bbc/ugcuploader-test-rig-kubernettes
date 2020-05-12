import React from 'react';
import Main from './Main/Main';
import MyProvider from './MyProvider';
import 'semantic-ui-css/semantic.min.css'
import './App.css';

function App() {
  return (
    <MyProvider>
      <Main/>
    </MyProvider>

  );
}

export default App;
