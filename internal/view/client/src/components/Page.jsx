import Container from '@mui/material/Container';
import Paper from '@mui/material/Paper';
import React from 'react';
import {node} from 'prop-types';

function Page({children}) {
  return (
    <Container maxWidth="lg">
      <Paper elevation={0} sx={{flexGrow: 1, marginTop: 5}}>
        {children}
      </Paper>
    </Container>
  );
}

Page.propTypes = {
  children: node.isRequired,
};

export default Page;
