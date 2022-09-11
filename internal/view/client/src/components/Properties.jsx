import Paper from '@mui/material/Paper';
import React from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';
import {object} from 'prop-types';

function formatValue(value) {
  if (typeof value === 'string') {
    return value;
  }
  if (typeof value === 'number') {
    return value;
  }
  return JSON.stringify(value);
}

function Properties({data}) {
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableBody>
          {Object.keys(data).map(key => (
            <TableRow key={key}>
              <TableCell>{key}</TableCell>
              <TableCell>{formatValue(data[key])}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

Properties.propTypes = {
  data: object.isRequired,
};

export default Properties;
