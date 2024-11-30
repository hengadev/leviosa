<script lang="ts">
    
    
    interface Props {
        placeholder: string | undefined;
        label: string;
        // TODO: change the type so that it can be a enum
        id: string;
        name: string;
        forInput: string;
        type: string;
        // TODO: remove that when the app goes to production
        value: string;
    }

    let {
        placeholder,
        label,
        id,
        name,
        forInput,
        type,
        value
    }: Props = $props();
    console.log("the type is:", type)
</script>

<div class="wrapper">
    <label for={forInput} class="label capitalize">{label}</label>
    <input 
        id={id}
        class="stroke" 
        name={name}
        type={type}
        placeholder={placeholder}
        value={value}
        required
    >
</div>

<style>
    .wrapper {
        position:relative;
        width: 100%;
    }

    .label {
        --paddinginline: 0.2rem;
        position: absolute;
        top: -0.75rem;
        top: -1rem;
        top: calc((-1 * var(--p) / 2) - 0.3rem);
        left: calc(1rem - var(--paddinginline));
        background-color: white;
        padding-inline: var(--paddinginline);
    }

    input {
        width: 100%;
        border-radius: 0.5rem;
        padding: var(--h3);
    }
    /* input:not(:placeholder-shown):valid { outline: 1px solid hsl(var(--clr-success)); } */
    input:not(:placeholder-shown):valid { outline: 1px solid hsl(var(--clr-dark-ternary)); }
    input:not(:placeholder-shown):invalid { outline: 1px solid hsl(var(--clr-error)); }
    input:focus:invalid { outline: 1px solid yellow; }

    /* ::placeholder, .label { color: hsl(var(--clr-stroke)); } */
    ::placeholder, .label { color: hsl(var(--clr-dark-ternary) / 0.8); }
</style>
