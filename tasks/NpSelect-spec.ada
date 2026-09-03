--  <vc-preamble>
package Np_Select_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Real  : constant := 1.0E6;

   subtype Index_Type is Natural range 0 .. Max_Index;

   subtype Real_Value is Long_Float range -Max_Real .. Max_Real;

   type Real_Array is array (Index_Type range <>) of Real_Value;

   --  Dafny's array<array<bool>> / array<array<real>>.  The outer index is
   --  the first dimension; the requires clauses force every inner array to
   --  have the same length, so a rectangular array is faithful.
   type Bool_Matrix is
     array (Index_Type range <>, Index_Type range <>) of Boolean;

   type Real_Matrix is
     array (Index_Type range <>, Index_Type range <>) of Real_Value;
--  </vc-preamble>

--  <vc-spec>
   function Select_Fn
     (Condlist : Bool_Matrix; Choicelist : Real_Matrix) return Real_Array
   with
     Pre  => Condlist'Length (1) > 0
             and then Condlist'Length (2) > 0
             and then Condlist'First (1) = Choicelist'First (1)
             and then Condlist'Last (1) = Choicelist'Last (1)
             and then Condlist'First (2) = Choicelist'First (2)
             and then Condlist'Last (2) = Choicelist'Last (2),
     Post => Select_Fn'Result'Length = Condlist'Length (2)
             and then Select_Fn'Result'First = Condlist'First (2)
             and then (for all I in Condlist'Range (1) =>
                         (for all J in Condlist'Range (2) =>
                            (if Condlist (I, J)
                             then Select_Fn'Result (J)
                                  = Choicelist (I, J))));

end Np_Select_Spec;

package body Np_Select_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Select_Fn
     (Condlist : Bool_Matrix; Choicelist : Real_Matrix) return Real_Array
   is
      Result : Real_Array (Condlist'Range (2)) := (others => 0.0);
   begin
      pragma Assume (False);
      return Result;
   end Select_Fn;
--  </vc-code>

--  <vc-postamble>
end Np_Select_Spec;
--  </vc-postamble>
