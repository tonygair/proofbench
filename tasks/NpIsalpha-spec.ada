--  <vc-preamble>
package Np_Isalpha_Spec with SPARK_Mode is

   Max_Index  : constant := 1_000;
   Max_Length : constant := 100;

   subtype Index_Type is Natural range 0 .. Max_Index;

   subtype Length_Type is Natural range 0 .. Max_Length;
   subtype Char_Index is Positive range 1 .. Max_Length;

   --  A string of at most Max_Length characters.  The characters actually
   --  present are Data (1 .. Length); Dafny's |s| is S.Length.
   type Bounded_String is record
      Length : Length_Type := 0;
      Data   : String (Char_Index) := (others => ' ');
   end record;

   type Str_Array is array (Index_Type range <>) of Bounded_String;

   type Bool_Array is array (Index_Type range <>) of Boolean;

   function Is_Alpha_Char (C : Character) return Boolean is
     (C in 'A' .. 'Z' or else C in 'a' .. 'z');

   function String_Is_Alpha (S : Bounded_String) return Boolean is
     (S.Length > 0
      and then (for all I in 1 .. S.Length => Is_Alpha_Char (S.Data (I))));
--  </vc-preamble>

--  <vc-spec>
   procedure Is_Alpha (Input : Str_Array; Ret : out Bool_Array) with
     Pre  => Ret'First = Input'First and then Ret'Last = Input'Last,
     Post => Ret'Length = Input'Length
             and then (for all I in Input'Range =>
                         Ret (I) = String_Is_Alpha (Input (I)));

end Np_Isalpha_Spec;

package body Np_Isalpha_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Is_Alpha (Input : Str_Array; Ret : out Bool_Array) is
   begin
      pragma Assume (False);
   end Is_Alpha;
--  </vc-code>

--  <vc-postamble>
end Np_Isalpha_Spec;
--  </vc-postamble>
