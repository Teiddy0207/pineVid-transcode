<template>
  <label :for="name">{{ label }}</label>
  <input :name="name" :id="name" :type="type" :value="inputValue" :placeholder="placeholder"
         @input="handleChange" @blur="handleBlur" class="form-control"/>
  <p class="invalid-feedback d-block" v-show="errorMessage || meta.valid">
    {{ errorMessage || successMessage }}
  </p>
</template>

<script>
  import {useField} from 'vee-validate'

  export default {
    props: {
      type: {
        type: String,
        default: 'text'
      },
      value: {
        type: String,
        default: ''
      },
      name: {
        type: String,
        required: true
      },
      label: {
        type: String,
        required: true
      },
      successMessage: {
        type: String,
        default: ''
      },
      placeholder: {
        type: String,
        default: ''
      }
    },
    setup(props) {
      const {
        value: inputValue,
        errorMessage,
        handleBlur,
        handleChange,
        meta
      } = useField(props.name, undefined, {
        initialValue: props.value
      })

      return {
        handleChange,
        handleBlur,
        errorMessage,
        inputValue,
        meta
      }
    }
  }
</script>
