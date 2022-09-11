import OLSTAC from 'ol/layer/STAC.js';
import {createElement, forwardRef} from 'react';
import {node} from 'prop-types';

const STAC = forwardRef(({children, ...props}, ref) => {
  return createElement('layer', {cls: OLSTAC, ref, ...props}, children);
});

STAC.propTypes = {
  children: node,
};

export default STAC;
